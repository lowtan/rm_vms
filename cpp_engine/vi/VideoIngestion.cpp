#include "VideoIngestion.h"

#include <iostream>
// #include <string>
#include <chrono>

extern "C" {
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h> // You may need this for av_packet_alloc, etc.
}

// --- Constructor ---
VideoIngestion::VideoIngestion(std::shared_ptr<ISharedMemory> mm, int id, const std::string u)
    : shm(mm), camID(id), url(u)
{
    camName = "[Cam" + std::to_string(camID) + "]";
    shmChannelID = shm->ChannelForCamID(camID);

    if(shmChannelID < 0) {

        Log::error(camName + "SHM reached max channel!");

    } else {

        workerThread = std::thread(&VideoIngestion::startIngestion, this);

    }

}

// --- Destructor ---
VideoIngestion::~VideoIngestion() {

    // Signal the thread to stop
    stopSignal = true;

    // Wake up the disk writer thread and tell it to exit safely
    diskWriterQueue.push(nullptr);

    // Join the disk writer thread
    if (diskWriterThread.joinable()) {
        diskWriterThread.join();
    }

    // Wait for the thread to finish (Join)
    // If we don't join, the thread might try to access 'this' after the object is destroyed -> Crash.
    if (workerThread.joinable()) {
        workerThread.join();
    }

}


/**
 * =========================================================
 * --- Private Method: startIngestion ---
 * =========================================================
 */
int VideoIngestion::startIngestion() {
    avformat_network_init();
    options = configureAVDictionary(nullptr);

    // Connect to Camera
    if (openInput() < 0) return cleanup();
    if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {
        Log::error(camName + " Could not retrieve stream info.");
        return cleanup();
    }

    // Locate Streams
    findStreamIndices();
    if (videoStreamIndex == -1) {
        Log::error(camName + " No video stream found.");
        return cleanup();
    }

    // Setup Filters
    if (initVideoFilter() < 0) return cleanup();

    // ---------------------------------------------------------
    // SPAWN WRITER THREAD HERE
    // Now that we have AVStreams, we can pass codec parameters to the writer
    // ---------------------------------------------------------
    initDiskWriter();

    Log::info(camName + " Connected! Starting Ingestion Loop...");
    Log::send("{\"status\":\"streaming\", \"cam\":" + std::to_string(camID) + ", \"channel\":" + std::to_string(shmChannelID) + "}");

    // The Main Loop
    AVPacket* packet = av_packet_alloc();
    while (!stopSignal) {
        if (av_read_frame(fmtCtx, packet) < 0) {
            Log::info(camName + " Error or End of Stream.");
            break; // Drop out of loop to trigger reconnect/cleanup
        }

        routePacket(packet);

        // Reset the packet for the next av_read_frame iteration
        av_packet_unref(packet); 
    }

    av_packet_free(&packet);
    return cleanup();
}


/**
 * =========================================================
 * Initializations
 * =========================================================
 */
void VideoIngestion::findStreamIndices() {
    videoStreamIndex = -1;
    audioStreamIndex = -1;

    for (unsigned int i = 0; i < fmtCtx->nb_streams; i++) {
        if (fmtCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
            videoStreamIndex = i;
        } else if (fmtCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_AUDIO) {
            audioStreamIndex = i;
        }
    }
}

void VideoIngestion::initDiskWriter() {

    diskWriterThread = std::thread([this]() {
        AVStream* vStream = (videoStreamIndex != -1) ? fmtCtx->streams[videoStreamIndex] : nullptr;
        AVStream* aStream = (audioStreamIndex != -1) ? fmtCtx->streams[audioStreamIndex] : nullptr;
        
        // Pass both streams to the worker
        writerWorker(this->diskWriterQueue, vStream, aStream, this->camID);
    });

}

int VideoIngestion::initVideoFilter() {
    const AVBitStreamFilter *bsf = av_bsf_get_by_name("dump_extra");
    
    if (av_bsf_alloc(bsf, &bsfCtx) < 0) {
        Log::error(camName + " Failed to allocate dump_extra BSF.");
        return -1;
    }

    if (avcodec_parameters_copy(bsfCtx->par_in, fmtCtx->streams[videoStreamIndex]->codecpar) < 0) {
        Log::error(camName + " Failed to copy parameters to BSF.");
        return -1;
    }

    if (av_bsf_init(bsfCtx) < 0) {
        Log::error(camName + " Failed to initialize BSF.");
        return -1;
    }

    return 0;
}

void VideoIngestion::routePacket(AVPacket* packet) {
    if (packet->stream_index == videoStreamIndex) {
        ingestVideo(packet);
    } else if (packet->stream_index == audioStreamIndex) {
        ingestAudio(packet);
    } 
    // If it's metadata or subtitles, we just do nothing.
    // The orchestrator loop will safely unref it.
}

/**
 * =========================================================
 * Disk Writer
 * =========================================================
 */

void VideoIngestion::packetToDiskWriter(AVPacket* packet) {

    AVPacket* cloneForDisk = av_packet_alloc();
    if (av_packet_ref(cloneForDisk, packet) >= 0) {

        // If the queue rejects the packet, WE must free the clone.
        if (!diskWriterQueue.push(cloneForDisk)) {
            av_packet_unref(cloneForDisk);
            av_packet_free(&cloneForDisk);
        }

    } else {
        av_packet_free(&cloneForDisk);
        Log::error(camName + "Failed to ref-count packet for disk queue.");
    }

}

/**
 * =========================================================
 * Ingestion
 * =========================================================
 */
void VideoIngestion::ingestVideo(AVPacket* packet) {
    // Send raw packet to the filter
    if (av_bsf_send_packet(bsfCtx, packet) == 0) {
        
        // Receive the modified packet(s) back
        while (av_bsf_receive_packet(bsfCtx, packet) == 0) {
            
            bool isKey = (packet->flags & AV_PKT_FLAG_KEY);

            if (waitForKeyFrame) {
                if (isKey) {
                    Log::info(camName + " [ingestVideo] First key frame found.");
                    waitForKeyFrame = false;
                } else {
                    av_packet_unref(packet); 
                    continue;
                }
            }

            try {

                if (shm->WriteFrame(shmChannelID, packet->data, packet->size, packet->pts, isKey) < 0) {
                    Log::error(camName + " [ingestVideo] Failed to write frame data.");
                }
                packetToDiskWriter(packet);

            } catch(...) {
                Log::error(camName + " [ingestVideo] Caught exception writing frame data.");
            }

            // Clean up the filtered packet
            av_packet_unref(packet);
        }
    }
}

void VideoIngestion::ingestAudio(AVPacket* packet) {

    // TODO: Implement audio extraction and routing to Audio SHM Buffer

    // A/V Sync Gatekeeper: Drop audio until video has established a keyframe.
    // This prevents "black screen with audio" at the start of recordings/streams.
    if (waitForKeyFrame) {
        // We haven't found a video I-Frame yet. Discard this audio packet.
        av_packet_unref(packet);
        return;
    }

    packetToDiskWriter(packet);

}


// --- Private Method: openInput ---
/**
 * Opens url and handles error message
 * @return 0 on success, -1 on failure.
 */
int VideoIngestion::openInput() {

    Log::info(camName + "Connecting to: " + url);

    int ret = avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options);
    if (ret != 0) {
        // Create a buffer for the error message
        char errbuf[256];
        
        // Ask FFmpeg to translate the error code
        av_strerror(ret, errbuf, sizeof(errbuf));

        std::cerr << camName << "[FFmpeg Error] Could not open source: " << url << std::endl;
        std::cerr << "Reason: " << errbuf << " (Code: " << ret << ")" << std::endl;
        // Log::send("{\"status\":\"stopped\", \"cam\":" + std::to_string(camID) + "}");

        return -1;
    }

    return 0;

}

int VideoIngestion::cleanup() {
    // Free the Bitstream Filter (Fixes the memory leak!)
    if (bsfCtx) {
        av_bsf_free(&bsfCtx);
        bsfCtx = nullptr;
    }

    // Close the input and free context
    if (fmtCtx) {
        avformat_close_input(&fmtCtx); 
        fmtCtx = nullptr;
    }

    // Free the dictionary options
    if (options) {
        av_dict_free(&options);
        options = nullptr;
    }

    // De-initialize network
    avformat_network_deinit();

    Log::info(camName + " Thread Exited cleanly.");
    Log::send("{\"status\":\"stopped\", \"cam\":" + std::to_string(camID) + "}");

    return -1; // Or return 0 depending on how your worker thread monitors exits
}
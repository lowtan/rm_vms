#include "VideoIngestion.h"

#include <iostream>
// #include <string>
#include <chrono>

#include "Log.h"

/**
 * Checks for any options that FFmpeg did NOT consume.
 * Useful for debugging typos (e.g. "timeout" vs "stimeout").
 * * @param dict The dictionary to check.
 * @param tag  Optional string to identify which stream this belongs to.
 */
void logUnusedOptions(AVDictionary* dict, const std::string& tag = "Stream") {
    AVDictionaryEntry* t = nullptr;
    
    // Iterate over all remaining entries in the dictionary
    while ((t = av_dict_get(dict, "", t, AV_DICT_IGNORE_SUFFIX))) {
        std::cerr << "[Warning][" << tag << "] Unused Option: Key='" << t->key 
                  << "', Value='" << t->value << "'" << std::endl;
    }
}

// --- Constructor ---
VideoIngestion::VideoIngestion(std::shared_ptr<ISharedMemory> mm, int id, const std::string u)
    : shm(mm), camID(id), url(u), stopSignal(false), fmtCtx(nullptr), options(nullptr)
{
    camName = "[Cam" + std::to_string(camID) + "]";
    workerThread = std::thread(&VideoIngestion::startIngestion, this);
}

// --- Destructor ---
VideoIngestion::~VideoIngestion() {

    // Signal the thread to stop
    stopSignal = true;

    // Wait for the thread to finish (Join)
    // If we don't join, the thread might try to access 'this' after the object is destroyed -> Crash.
    if (workerThread.joinable()) {
        workerThread.join();
    }

}



// --- Private Method: startIngestion ---
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

    Log::info(camName + " Connected! Starting Ingestion Loop...");
    Log::send("{\"status\":\"starting\", \"cam\":" + std::to_string(camID) + "}");

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
                    Log::info(camName + " [SHM] First key frame found.");
                    waitForKeyFrame = false;
                } else {
                    av_packet_unref(packet); 
                    continue;
                }
            }

            try {
                if (shm->WriteFrame(camID, packet->data, packet->size, packet->pts, isKey) < 0) {
                    Log::error(camName + " [SHM] Failed to write frame data.");
                }
            } catch(...) {
                Log::error(camName + " [SHM] Caught exception writing frame data.");
            }

            // Clean up the filtered packet
            av_packet_unref(packet);
        }
    }
}

void VideoIngestion::ingestAudio(AVPacket* packet) {
    // TODO: Implement audio extraction and routing to Audio SHM Buffer
}

// int VideoIngestion::startIngestion() {

//     // Initialize FFmpeg Network
//     avformat_network_init();

//     options = configureAVDictionary(nullptr);

//     if(this->openInput()==0) {

//         if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {
//             Log::error(camName + "Could not retrieve stream info.");
//             // 
//         } else {

//             // Locate the Video Stream Index
//             int videoStreamIndex = -1;
//             for (unsigned int i = 0; i < fmtCtx->nb_streams; i++) {
//                 if (fmtCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
//                     videoStreamIndex = i;
//                     break;
//                 }
//             }

//             if (videoStreamIndex == -1) {

//                 std::cerr << camName << "No video stream found." << std::endl;

//             } else {

//                 Log::info(camName + "Connected! Starting Ingestion Loop...");
//                 Log::send("{CamID: " + std::to_string(camID) + ", Status: 1}");

//                 // --- INITIALIZE BITSTREAM FILTER ---
//                 const AVBitStreamFilter *bsf = av_bsf_get_by_name("dump_extra");
//                 AVBSFContext *bsfCtx = nullptr;

//                 if (av_bsf_alloc(bsf, &bsfCtx) < 0) {
//                     Log::error(camName + " Failed to allocate dump_extra BSF.");
//                     return -1;
//                 }

//                 // Copy the camera's codec parameters (containing the SPS/PPS) into the filter
//                 if (avcodec_parameters_copy(bsfCtx->par_in, fmtCtx->streams[videoStreamIndex]->codecpar) < 0) {
//                     Log::error(camName + " Failed to copy parameters to BSF.");
//                     av_bsf_free(&bsfCtx);
//                     return -1;
//                 }

//                 if (av_bsf_init(bsfCtx) < 0) {
//                     Log::error(camName + " Failed to initialize BSF.");
//                     av_bsf_free(&bsfCtx);
//                     return -1;
//                 }
//                 // -----------------------------------


//                 bool waitForKeyFrame = true;

//                 // Allocation: Create a packet container
//                 // An AVPacket holds the compressed data (e.g., one H.264 chunk)
//                 AVPacket* packet = av_packet_alloc();

//                 // Log::info("check stopSignal" + std::to_string(stopSignal));
//                 // --- THE CRITICAL LOOP ---
//                 while (!stopSignal) {

//                     // == Read a frame from the network
//                     // av_read_frame grabs the next RTP packet(s) and assembles them 
//                     // into a single "Access Unit" (compressed frame)
//                     int ret = av_read_frame(fmtCtx, packet);

//                     if (ret < 0) {
//                         std::cerr << camName << "Error or End of Stream." << std::endl;
//                         break; // Reconnect logic would go here
//                     }

//                     // == Filter: Only process packets belonging to the video stream
//                     // == Filter: Only process packets belonging to the video stream
//                     if (packet->stream_index == videoStreamIndex) {

//                         // Send raw packet to the filter 
//                         // (This consumes the original packet reference on success)
//                         if (av_bsf_send_packet(bsfCtx, packet) == 0) {
                            
//                             // Receive the modified packet(s) back
//                             while (av_bsf_receive_packet(bsfCtx, packet) == 0) {
                                
//                                 bool isKey = (packet->flags & AV_PKT_FLAG_KEY);

//                                 if (waitForKeyFrame) {
//                                     if (isKey) {
//                                         Log::info(camName + "[SHM] First key frame found. " + std::to_string(camID));
//                                         waitForKeyFrame = false;
//                                     } else {
//                                         // MUST clean up the packet before skipping
//                                         av_packet_unref(packet); 
//                                         continue;
//                                     }
//                                 }

//                                 try {
//                                     // Push the newly filtered packet to Ring Buffer
//                                     if (shm->WriteFrame(camID, packet->data, packet->size, packet->pts, isKey) < 0) {
//                                         Log::error(camName + "[SHM] Failed to write frame data for cam:" + std::to_string(camID));
//                                     }
//                                 } catch(...) {
//                                     Log::error(camName + "[SHM] Catched, Failed to write frame data for cam:" + std::to_string(camID));
//                                 }

//                                 // Clean up the filtered packet after writing
//                                 av_packet_unref(packet);
//                             }
//                         } else {
//                             // If send fails, clean up the original packet
//                             av_packet_unref(packet);
//                         }
//                     } else {
//                         // Not a video stream packet, discard it
//                         av_packet_unref(packet);
//                     }


//                 }
//                 // --- END: THE CRITICAL LOOP ---

//                 av_packet_free(&packet);

//             }

//         }

//     }

//     // Close the input. This frees fmtCtx and closes the socket.
//     if (fmtCtx) {
//         avformat_close_input(&fmtCtx); 
//         fmtCtx = nullptr;
//     }

//     // Free the dictionary options
//     if (options) {
//         av_dict_free(&options);
//         options = nullptr;
//     }

//     // Deinit network
//     avformat_network_deinit();

//     Log::info(camName + " Thread Exited.");
//     Log::send("{\"status\":\"stopped\", \"cam\":" + std::to_string(camID) + "}");

//     return 0;
// }

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
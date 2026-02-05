#include "VideoIngestion.h"

#include <iostream>
// #include <string>
#include <thread>
#include <chrono>
#include <atomic>

extern "C" {
#include <libavformat/avformat.h>
}


#include "Log.h"
#include "AVDictionary.h"

// A flag to stop the loop cleanly (e.g., from a signal handler)
std::atomic<bool> keepRunning(true);


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

int startIngestion(int camID, const std::string& url) {

    // Initialize FFmpeg Network
    avformat_network_init();

    // In a real app, you would spawn a std::thread here.
    // For this skeleton, we just mock the connection check.
    AVFormatContext* fmtCtx = nullptr;
    AVDictionary* options = configureAVDictionary(nullptr);

    Log::info("Connecting to: " + url);

    if (avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options) == 0) {

        Log::info("Connection Successful!");

        logUnusedOptions(options, std::to_string(camID));

        // Find Stream Info (Parses the header to find Video/Audio)
        if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {

            std::cerr << "Could not retrieve stream info." << std::endl;

            avformat_free_context(fmtCtx);
            av_dict_free(&options);
            return -1;

        } else {

            // Locate the Video Stream Index
            int videoStreamIndex = -1;
            for (unsigned int i = 0; i < fmtCtx->nb_streams; i++) {
                if (fmtCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
                    videoStreamIndex = i;
                    break;
                }
            }

            if (videoStreamIndex == -1) {
                std::cerr << "No video stream found." << std::endl;
                return -1;
            }

            std::cout << "Connected! Starting Ingestion Loop..." << std::endl;
            std::cout << "{CamID: " << camID << ", Status: 1}" << std::endl;

            // Allocation: Create a packet container
            // An AVPacket holds the compressed data (e.g., one H.264 chunk)
            AVPacket* packet = av_packet_alloc();

            // --- THE CRITICAL LOOP ---
            while (keepRunning) {

                // == Read a frame from the network
                // av_read_frame grabs the next RTP packet(s) and assembles them 
                // into a single "Access Unit" (compressed frame)
                int ret = av_read_frame(fmtCtx, packet);
                
                if (ret < 0) {
                    std::cerr << "Error or End of Stream." << std::endl;
                    break; // Reconnect logic would go here
                }

                // == Filter: Only process packets belonging to the video stream
                if (packet->stream_index == videoStreamIndex) {
                    
                    // Check if it's a Keyframe (I-Frame)
                    if (packet->flags & AV_PKT_FLAG_KEY) {
                        std::cout << "Found Keyframe! Size: " << packet->size << std::endl;
                    }

                    // Push to Ring Buffer
                    // PushToSharedMemory(packet->data, packet->size, packet->pts);

                    // Write to Disk (Muxing)
                    // WriteToFile(packet);
                }

                // Clean up: we MUST wipe the packet after using it,
                // otherwise memory will explode in seconds.
                av_packet_unref(packet);
            }

            // Cleanup
            av_packet_free(&packet);
            avformat_network_deinit();

        }

        av_dict_free(&options);
        avformat_close_input(&fmtCtx);

        return 0;

    }

    Log::error("Connection Failed (Expected if URL is fake)");
    return -1;

}

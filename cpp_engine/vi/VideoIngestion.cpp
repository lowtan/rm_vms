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
VideoIngestion::VideoIngestion(int id, const std::string u)
    : camID(id), url(u), stopSignal(false), fmtCtx(nullptr), options(nullptr)
{
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

    // Initialize FFmpeg Network
    avformat_network_init();

    options = configureAVDictionary(nullptr);

    if(this->openInput()==0) {

        if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {
            Log::error("Could not retrieve stream info.");
            // 
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

            } else {


                Log::info("Connected! Starting Ingestion Loop...");
                Log::info("{CamID: " + std::to_string(camID) + ", Status: 1}");

                // Allocation: Create a packet container
                // An AVPacket holds the compressed data (e.g., one H.264 chunk)
                AVPacket* packet = av_packet_alloc();

                // --- THE CRITICAL LOOP ---
                while (!stopSignal) {

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

                    // -- Essential per-frame cleanup --
                    // Clean up: we MUST wipe the packet after using it,
                    // otherwise memory will explode in seconds.
                    av_packet_unref(packet);

                }
                // --- END: THE CRITICAL LOOP ---

                av_packet_free(&packet);

            }

        }

    }

    // Close the input. This frees fmtCtx and closes the socket.
    if (fmtCtx) {
        avformat_close_input(&fmtCtx); 
        fmtCtx = nullptr;
    }

    // 2. Free the dictionary options
    if (options) {
        av_dict_free(&options);
        options = nullptr;
    }

    // 3. Deinit network
    avformat_network_deinit();

    Log::info("[Cam " + std::to_string(camID) + "] Thread Exited.");

    return 0;
}

// --- Private Method: openInput ---
/**
 * Opens url and handles error message
 * @return 0 on success, -1 on failure.
 */
int VideoIngestion::openInput() {

    Log::info("Connecting to: " + url);

    int ret = avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options);
    if (ret != 0) {
        // 1. Create a buffer for the error message
        char errbuf[256];
        
        // 2. Ask FFmpeg to translate the error code
        av_strerror(ret, errbuf, sizeof(errbuf));

        // 3. Print it
        std::cerr << "[FFmpeg Error] Could not open source: " << url << std::endl;
        std::cerr << "Reason: " << errbuf << " (Code: " << ret << ")" << std::endl;

        return -1;
    }

    return 0;

}



// int startIngestion(int camID, const std::string& url) {

//     // Initialize FFmpeg Network
//     avformat_network_init();

//     // In a real app, you would spawn a std::thread here.
//     // For this skeleton, we just mock the connection check.
//     AVFormatContext* fmtCtx = nullptr;
//     AVDictionary* options = configureAVDictionary(nullptr);

//     Log::info("Connecting to: " + url);

//     if (avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options) == 0) {

//         Log::info("Connection Successful!");

//         logUnusedOptions(options, std::to_string(camID));

//         // Find Stream Info (Parses the header to find Video/Audio)
//         if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {

//             Log::error("Could not retrieve stream info.");

//             avformat_free_context(fmtCtx);
//             av_dict_free(&options);
//             return -1;

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
//                 std::cerr << "No video stream found." << std::endl;
//                 return -1;
//             }

//             std::cout << "Connected! Starting Ingestion Loop..." << std::endl;
//             std::cout << "{CamID: " << camID << ", Status: 1}" << std::endl;

//             // Allocation: Create a packet container
//             // An AVPacket holds the compressed data (e.g., one H.264 chunk)
//             AVPacket* packet = av_packet_alloc();

//             // --- THE CRITICAL LOOP ---
//             while (keepRunning) {

//                 // == Read a frame from the network
//                 // av_read_frame grabs the next RTP packet(s) and assembles them 
//                 // into a single "Access Unit" (compressed frame)
//                 int ret = av_read_frame(fmtCtx, packet);

//                 if (ret < 0) {
//                     std::cerr << "Error or End of Stream." << std::endl;
//                     break; // Reconnect logic would go here
//                 }

//                 // == Filter: Only process packets belonging to the video stream
//                 if (packet->stream_index == videoStreamIndex) {
                    
//                     // Check if it's a Keyframe (I-Frame)
//                     if (packet->flags & AV_PKT_FLAG_KEY) {
//                         std::cout << "Found Keyframe! Size: " << packet->size << std::endl;
//                     }

//                     // Push to Ring Buffer
//                     // PushToSharedMemory(packet->data, packet->size, packet->pts);

//                     // Write to Disk (Muxing)
//                     // WriteToFile(packet);
//                 }

//                 // Clean up: we MUST wipe the packet after using it,
//                 // otherwise memory will explode in seconds.
//                 av_packet_unref(packet);
//             }

//             // Cleanup
//             av_packet_free(&packet);
//             avformat_network_deinit();

//         }

//         av_dict_free(&options);
//         avformat_close_input(&fmtCtx);

//         return 0;

//     }

//     Log::error("Connection Failed (Expected if URL is fake)");
//     return -1;

// }

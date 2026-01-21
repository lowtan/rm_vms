#include <iostream>
#include <string>
#include <thread>
#include <chrono>
#include <atomic>

extern "C" {
#include <libavformat/avformat.h>
}

// A flag to stop the loop cleanly (e.g., from a signal handler)
std::atomic<bool> keepRunning(true);

// Simple logging helper to Stderr (so it doesn't mess up JSON on Stdout)
void log(const std::string& msg) {
    std::cerr << "[C++ Worker] " << msg << std::endl;
}

int startIngestion(int camID, const std::string& url) {
    // 1. Initialize FFmpeg Network
    avformat_network_init();
    
    // In a real app, you would spawn a std::thread here.
    // For this skeleton, we just mock the connection check.
    
    AVFormatContext* fmtCtx = nullptr;
    AVDictionary* options = nullptr;
    av_dict_set(&options, "rtsp_transport", "tcp", 0); // Force TCP
    av_dict_set(&options, "stimeout", "2000000", 0);   // 2s timeout

    log("Connecting to: " + url);

    if (avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options) == 0) {

        log("Connection Successful!");

        // Find Stream Info (Parses the header to find Video vs Audio)
        if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {
            std::cerr << "Could not retrieve stream info." << std::endl;
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

            // 7. Allocation: Create a packet container
            // An AVPacket holds the compressed data (e.g., one H.264 chunk)
            AVPacket* packet = av_packet_alloc();

            // int keepRunning = 1;

            // --- THE CRITICAL LOOP ---
            while (keepRunning) {
                // A. Read a frame from the network
                // av_read_frame grabs the next RTP packet(s) and assembles them 
                // into a single "Access Unit" (compressed frame)
                int ret = av_read_frame(fmtCtx, packet);
                
                if (ret < 0) {
                    std::cerr << "Error or End of Stream." << std::endl;
                    break; // Reconnect logic would go here
                }

                // B. Filter: Only process packets belonging to the video stream
                if (packet->stream_index == videoStreamIndex) {
                    
                    // --- YOUR CUSTOM LOGIC GOES HERE ---
                    
                    // Example 1: Check if it's a Keyframe (I-Frame)
                    if (packet->flags & AV_PKT_FLAG_KEY) {
                        std::cout << "Found Keyframe! Size: " << packet->size << std::endl;
                    }

                    // Example 2: Push to Ring Buffer
                    // PushToSharedMemory(packet->data, packet->size, packet->pts);
                    
                    // Example 3: Write to Disk (Muxing)
                    // WriteToFile(packet);
                }

                // C. Clean up: You MUST wipe the packet after using it, 
                // otherwise memory will explode in seconds.
                av_packet_unref(packet);
            }

            // Cleanup
            av_packet_free(&packet);
            avformat_network_deinit();

        }


        avformat_close_input(&fmtCtx);

    } else {
        log("Connection Failed (Expected if URL is fake)");
    }
}

int main() {
    // Optimize I/O
    std::ios_base::sync_with_stdio(false);
    std::cin.tie(NULL);

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line == "EXIT") break;

        // Basic parsing: "START <ID> <URL>"
        if (line.substr(0, 5) == "START") {
            try {
                size_t firstSpace = line.find(' ');
                size_t secondSpace = line.find(' ', firstSpace + 1);
                
                std::string idStr = line.substr(firstSpace + 1, secondSpace - firstSpace - 1);
                std::string url = line.substr(secondSpace + 1);
                
                // Respond to Go: "I received your command"
                std::cout << "{\"status\":\"starting\", \"cam\":" << idStr << "}" << std::endl;
                
                // Run logic
                startIngestion(std::stoi(idStr), url);

            } catch (...) {
                log("Error parsing command");
            }
        }
    }
    return 0;
}
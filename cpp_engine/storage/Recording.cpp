#include "Recording.h"



void Recording(AVPacket* packet) {



}

// writerWorker
// 
void writerWorker(SafeQueue<AVPacket*>& queue, AVCodecParameters* codecParams, AVRational inputTimeBase) {
    SegmentRecorder recorder;

    // Start the first file
    recorder.StartSegment("/storage/cam01/video_001.mkv", codecParams);

    auto lastSwitchTime = std::chrono::steady_clock::now();

    while (running) {
        AVPacket* packet = queue.pop();

        // Check if we need to rotate files (e.g., every 1 mins)
        auto now = std::chrono::steady_clock::now();
        if (std::chrono::duration_cast<std::chrono::minutes>(now - lastSwitchTime).count() >= 1) {

            // Only switch on a KEYFRAME (I-Frame)
            // If you switch on a P-frame, the next file will start with gray artifacts.
            if (packet->flags & AV_PKT_FLAG_KEY) {
                recorder.StopSegment();
                // Generate new filename with timestamp
                std::string newFile = GenerateFilename("cam01"); 
                recorder.StartSegment(newFile, codecParams);
                lastSwitchTime = now;
            }
        }

        // Write
        recorder.WritePacket(packet, inputTimeBase);

        // Clean up
        av_packet_free(&packet);
    }
    
    recorder.StopSegment();
}
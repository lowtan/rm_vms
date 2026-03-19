#include "Recording.h"

#include "SegmentRecorder.h"
#include "StorePath.h"
#include <chrono>
#include <string>

extern "C" {
#include <libavformat/avformat.h>
}

const int RECORDING_ROTATE_TIME = 1; // By minute

void writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID) {
    SegmentRecorder recorder;
    StorePath pathGenerator; // Reads from default "/recordings" internally
    
    auto lastSwitchTime = std::chrono::steady_clock::now();
    bool isFirstSegment = true;

    // The loop runs infinitely until the destructor pushes a nullptr
    while (true) {
        AVPacket* packet = queue.pop();

        // Graceful shutdown signal received from VideoIngestion teardown
        if (!packet) {
            break; 
        }

        auto now = std::chrono::steady_clock::now();
        bool timeToRotate = std::chrono::duration_cast<std::chrono::minutes>(now - lastSwitchTime).count() >= RECORDING_ROTATE_TIME;
        // bool timeToRotate = std::chrono::duration_cast<std::chrono::seconds>(now - lastSwitchTime).count() >= 30;

        // CRITICAL A/V FIX: Ensure we only rotate files on a VIDEO Keyframe.
        // Audio streams often mark every packet as a keyframe.
        bool isVideoPacket = (inVideoStream && packet->stream_index == inVideoStream->index);
        bool isVideoKeyframe = isVideoPacket && (packet->flags & AV_PKT_FLAG_KEY);

        // Start the first file or rotate with Video Keyframe boundary
        if (isFirstSegment || (timeToRotate && isVideoKeyframe)) {
            recorder.StopSegment();

            // Generate the precise directory tree and filename (e.g., /recordings/cam01/2026/03/12/15-30-00.mp4)
            std::string newFile = pathGenerator.For(camID, packet); 

            // Pass both streams to the SegmentRecorder so it can allocate the MP4 tracks
            recorder.StartSegment(newFile, inVideoStream, inAudioStream);

            lastSwitchTime = now;
            isFirstSegment = false;
        }

        // Delegate A/V routing, rescaling, and interleaving to the recorder
        recorder.WritePacket(packet);

        // Free the cloned packet memory allocated by av_packet_ref
        av_packet_unref(packet);
        av_packet_free(&packet);
    }

    // Finalize the last MP4 file to ensure the 'moov' atom is written before exiting
    recorder.StopSegment();
}
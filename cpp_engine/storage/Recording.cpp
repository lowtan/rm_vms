#include "Recording.h"

#include "SegmentRecorder.h"
#include "StorePath.h"
#include "Log.h"

#include <chrono>
#include <string>
#include <filesystem>

extern "C" {
#include <libavformat/avformat.h>
}

const int RECORDING_ROTATE_TIME = 1; // By minute


static inline long getCurrentUnixTime() {
    return std::chrono::duration_cast<std::chrono::seconds>(
        std::chrono::system_clock::now().time_since_epoch()).count();
}

RecorderWorker::RecorderWorker(std::string rp) : rootPath(rp) {}

long RecorderWorker::getEndTimeUnix(SegmentRecorder& recorder) {
    double duration = recorder.GetVideoDurationSeconds();
    return currentStartTimeUnix + static_cast<long>(duration);
}

// --- IPC Helper Function ---
void RecorderWorker::sendSegmentDoneIPC(int camID, long startTimeUnix, long endTimeUnix, const std::string& filePath) {
    // If there is no file path (e.g., the very first startup iteration), do nothing.
    if (filePath.empty()) return;

    // Get exact bytes written to the physical disk
    long sizeBytes = 0;
    std::error_code ec;
    sizeBytes = std::filesystem::file_size(filePath, ec);
    if (ec) sizeBytes = 0; // Failsafe if the file didn't write correctly

    // Construct and send JSON down the STDOUT pipe
    std::string json = "{\"status\":\"segment_done\", "
                       "\"cam\":" + std::to_string(camID) + ", "
                       "\"start_time\":" + std::to_string(startTimeUnix) + ", "
                       "\"end_time\":" + std::to_string(endTimeUnix) + ", "
                       "\"file_path\":\"" + filePath + "\", "
                       "\"size_bytes\":" + std::to_string(sizeBytes) + "}";

    Log::send(json);
}

void RecorderWorker::writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID) {
    StorePath pathGenerator;

    if(!rootPath.empty()) {
        pathGenerator = StorePath(rootPath);
    }

    auto lastSwitchTime = std::chrono::steady_clock::now();
    bool isFirstSegment = true;

    // std::string currentFilePath;
    long endTimeUnix;

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

            endTimeUnix = getEndTimeUnix(recorder);
            recorder.StopSegment();

            // Fire the IPC event for the completed file (Safely ignores the first run)
            sendSegmentDoneIPC(camID, currentStartTimeUnix, endTimeUnix, currentFilePath);

            // Generate the precise directory tree and filename (e.g., /recordings/cam01/2026/03/12/15-30-00.mp4)
            currentFilePath = pathGenerator.For(camID, packet); 
            currentStartTimeUnix = getCurrentUnixTime();

            // Pass both streams to the SegmentRecorder so it can allocate the MP4 tracks
            recorder.StartSegment(currentFilePath, inVideoStream, inAudioStream);

            lastSwitchTime = now;
            isFirstSegment = false;
        }

        // Delegate A/V routing, rescaling, and interleaving to the recorder
        recorder.WritePacket(packet);

        // Free the cloned packet memory allocated by av_packet_ref
        av_packet_unref(packet);
        av_packet_free(&packet);
    }

    endTimeUnix = getEndTimeUnix(recorder);
    recorder.StopSegment();

    // Fire the IPC event for latest file
    sendSegmentDoneIPC(camID, currentStartTimeUnix, endTimeUnix, currentFilePath);

}

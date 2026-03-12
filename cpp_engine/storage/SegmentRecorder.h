#pragma once

#include <string>

extern "C" {
#include <libavformat/avformat.h>
}

class SegmentRecorder {
private:
    AVFormatContext* outFormatCtx = nullptr;
    std::string currentFilename;
    bool isRecording = false;

    // Output stream indices in the MP4 file
    int outVideoStreamIndex = -1;
    int outAudioStreamIndex = -1;

    // Original input stream indices from the camera
    int inVideoStreamIndex = -1;
    int inAudioStreamIndex = -1;

    // Timebases for proper PTS/DTS rescaling
    AVRational videoInputTimeBase;
    AVRational audioInputTimeBase;

public:
    SegmentRecorder() = default;
    ~SegmentRecorder();

    // Now accepts both streams. Audio can be nullptr if the camera doesn't have a mic.
    bool StartSegment(const std::string& filename, AVStream* inVideoStream, AVStream* inAudioStream);

    void WritePacket(AVPacket* packet);
    void StopSegment();

    inline bool IsRecording() const { return isRecording; }
};
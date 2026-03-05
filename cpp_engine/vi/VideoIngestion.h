#pragma once
#include <string>
#include <thread>
#include <atomic>

#include "AVDictionary.h"
#include "SharedMemory.h"

extern "C" {
#include <libavformat/avformat.h>
}


class VideoIngestion
{
public:
    VideoIngestion(std::shared_ptr<ISharedMemory> mm, int id, const std::string u);
    ~VideoIngestion();

private:

    std::shared_ptr<ISharedMemory> shm;

    int camID;
    std::string camName;
    std::string url;

    // --- FFmpeg Contexts & Options ---
    AVFormatContext* fmtCtx;
    AVDictionary* options;
    AVBSFContext* bsfCtx;       // Bitstream filter context for SPS/PPS injection

    // Threading controls
    std::atomic<bool> stopSignal; // Each camera has its own stop flag
    std::thread workerThread;            // The thread handling the loop

    // --- Stream Tracking ---
    int videoStreamIndex;
    int audioStreamIndex;
    bool waitForKeyFrame;       // Ensures we drop P-frames until our first IDR

    int startIngestion();       // The main worker thread loop
    int openInput();            // Connects to the RTSP source
    int cleanup();              // Safely frees all FFmpeg resources

    // --- Setup Helpers ---
    void findStreamIndices();
    int initVideoFilter();

    // --- Packet Routing & Processing ---
    void routePacket(AVPacket* packet);
    void ingestVideo(AVPacket* packet);
    void ingestAudio(AVPacket* packet);
};

#pragma once
#include <string>
#include <thread>
#include <atomic>

#include "AVDictionary.h"
#include "SharedMemory.h"

extern "C" {
#include <libavformat/avformat.h>
}

// int startIngestion(int camID, const std::string& url);

class VideoIngestion
{
private:
    std::shared_ptr<ISharedMemory> shm;

    int camID;
    std::string camName;
    std::string url;

    // These vars should be initialized at startIngestion()
    AVDictionary* options;
    AVFormatContext* fmtCtx;

    // Threading controls
    std::atomic<bool> stopSignal; // Each camera has its own stop flag
    std::thread workerThread;            // The thread handling the loop

    int startIngestion();
    int openInput();

public:
    VideoIngestion(std::shared_ptr<ISharedMemory> mm, int id, const std::string u);
    ~VideoIngestion();
};

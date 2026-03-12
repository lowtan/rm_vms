#pragma once

#include "SafeQueue.h"

struct AVPacket;
struct AVStream;

// void Recording(AVPacket* packet);

// Spawns the worker loop for multiplexing packets to disk
void writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID);
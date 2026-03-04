#pragma once
extern "C" {
#include <libavformat/avformat.h>
}

class StorePath {

public:
    std::string& For(int camID, AVPacket* packet);
};
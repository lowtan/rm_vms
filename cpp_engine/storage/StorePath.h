#pragma once

#include <string>

struct AVPacket;

class StorePath {
private:
    std::string rootPath;

public:
    // Allow injecting the root path via constructor (e.g., from config.json later)
    StorePath(const std::string& root = "./recordings");

    // FIXED: Return by value to prevent dangling references.
    // Generates path: /recordings/cam01/YYYY/MM/DD/HH-MM-SS.mp4
    std::string For(int camID, AVPacket* packet);
};
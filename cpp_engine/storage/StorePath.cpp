#include "StorePath.h"
#include <iostream>
#include <sstream>
#include <iomanip>
#include <chrono>
#include <filesystem> // Requires C++17
#include <ctime>

extern "C" {
#include <libavformat/avformat.h>
}

StorePath::StorePath(const std::string& root) : rootPath(root) {}

std::string StorePath::For(int camID, AVPacket* packet) {

    //  Capture the exact wall-clock time the segment starts
    auto now = std::chrono::system_clock::now();
    auto in_time_t = std::chrono::system_clock::to_time_t(now);
    
    std::tm bt{};
    
    // POSIX thread-safe localtime. Perfect for macOS and Linux.
    localtime_r(&in_time_t, &bt);

    //  Construct the directory path: ROOT/camID/YYYY/MM/DD
    std::ostringstream folderStream;
    folderStream << rootPath 
                 << "/cam" << std::setfill('0') << std::setw(2) << camID 
                 << "/" << std::put_time(&bt, "%Y/%m/%d");

    std::string folderPath = folderStream.str();

    //  Ensure the directory structure exists (equivalent to `mkdir -p`)
    std::error_code ec;
    std::filesystem::create_directories(folderPath, ec);
    if (ec) {
        // In a production system, you might want to log this via your Log class
        std::cerr << "[StorePath] Critical IO Error: Failed to create directories: " 
                  << ec.message() << std::endl;
    }

    //  Construct the final filename: HH-MM-SS.mp4
    std::ostringstream fileStream;
    fileStream << folderPath << "/" << std::put_time(&bt, "%H-%M-%S") << ".mp4";

    return fileStream.str();
}
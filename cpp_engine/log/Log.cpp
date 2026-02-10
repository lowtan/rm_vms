#include "Log.h"
#include <iostream>

// // Simple logging helper to Stderr (so it doesn't mess up JSON on Stdout)
// void log(const std::string& msg) {
//     std::cout << "[cpp_engine] " << msg << std::endl;
// }

// void logErr(const std::string& msg) {
//     std::cerr << "[cpp_engine][err] " << msg << std::endl;
// }

void Log::info(const std::string& msg) {

    // Acquire the lock
    // Only one thread can pass this line at a time. 
    // Others will wait here until the lock is released.
    std::lock_guard<std::mutex> lock(s_Mutex);

    std::cout << "[cpp_engine] " << msg << std::endl;
}

void Log::error(const std::string& msg) {
    std::lock_guard<std::mutex> lock(s_Mutex);
    std::cerr << "[cpp_engine][err] " << msg << std::endl;
}
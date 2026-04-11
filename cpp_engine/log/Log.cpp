#include "Log.h"
#include <iostream>
#include <cstdarg> // Required for va_list, va_start
#include <vector>
#include <cstring> // for strerror

// Initialize the mutex
std::mutex Log::s_Mutex;

// Send data to parent Go program.
void Log::send(const std::string& msg) {

    // Acquire the lock
    // Only one thread can pass this line at a time. 
    // Others will wait here until the lock is released.
    std::lock_guard<std::mutex> lock(s_Mutex);

    std::cout << msg << std::endl;
}

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

// This method does not work, since Go won't print as is.
void Log::progress(const std::string& msg) {
    std::lock_guard<std::mutex> lock(s_Mutex);

    // \r moves cursor to the start.
    // \033[K is an ANSI escape code that clears the line from the cursor to the end.
    // std::flush forces it to draw immediately.
    std::cout << "\r\033[K " << msg << std::flush;
}

// "Printf-style" version
void Log::error(const char* format, ...) {
    // A fixed buffer is usually sufficient for log lines (fast stack allocation)
    char buffer[2048]; 

    // Initialize the variable argument list
    va_list args;
    va_start(args, format);

    // Format the string into the buffer
    // vsnprintf protects against buffer overflow
    vsnprintf(buffer, sizeof(buffer), format, args);

    // Clean up the list
    va_end(args);

    Log::error(buffer);
    // // Reuse the thread-safe logic
    // // We can just call the other overload or print directly
    // std::lock_guard<std::mutex> lock(s_Mutex);
    // std::cerr << "[cpp_engine][err] " << buffer << std::endl;
}

// Replacement for standard perror
void Log::PError(const std::string& prefix) {
    // buffer for the error message
    char errBuf[256]; 
    
    // standard, thread-safe way to get the error string
    // XSI-compliant strerror_r returns int
    // GNU-specific strerror_r returns char*
    // We use the safer, portable C++11 approach or standard POSIX:
    
    #if (_POSIX_C_SOURCE >= 200112L) && !  _GNU_SOURCE
        strerror_r(errno, errBuf, sizeof(errBuf));
        // Now log it using YOUR standard formatting
        Log::error("%s: %s", prefix.c_str(), errBuf); 
    #else
        // Fallback or GNU specific if needed, but often just:
        strerror_r(errno, errBuf, sizeof(errBuf));
        Log::error("%s: %s", prefix.c_str(), errBuf);
    #endif
}

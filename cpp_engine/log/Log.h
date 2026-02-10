#pragma once
#include <string>
#include <mutex>

/**
 * Besides to logging, Log is also used as a way to
 * communicate with Golang. So be aware of this when
 * using it.
 */

// void log(const std::string& msg);

class Log {
public:
    static void info(const std::string& msg);
    static void error(const std::string& msg);

private:
    // A single shared lock for the entire application
    static std::mutex s_Mutex;
};
#pragma once
#include <cstdint>

namespace utils {
    long getCurrentUnixTime();
    // Returns the current system wall-clock time in milliseconds since Unix Epoch
    uint64_t getCurrentEpochMSTime();
}
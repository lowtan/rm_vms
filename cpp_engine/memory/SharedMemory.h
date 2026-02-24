#pragma once
#include <string>
#include <memory>
#include <cstdint>
#include <atomic>

// Align to 64 bytes to prevent "False Sharing" on CPU cache lines
#define CACHE_LINE 64

struct FrameMetadata {
    uint32_t frameSize;    // Size of the actual video data
    uint64_t timestamp;    // Unix timestamp or PTS
    uint8_t  isKeyFrame;   // 1 = I-Frame, 0 = P/B-Frame
    uint8_t  _padding[51]; // Pad to 64 bytes for alignment
};

struct RingBufferHeader {
    std::atomic<uint32_t> writeHead;
    std::atomic<uint32_t> readTail;
    uint32_t bufferCapacity; // Renamed from bufferSize for clarity
    uint32_t streamID;
    uint8_t  _padding[48];
};

// Calculation:
// Buffer Base Address + sizeof(RingBufferHeader) = Start of Data

std::string ringBufferNameFor(std::string worker);

// ISharedMemory Interface
class ISharedMemory {
public:
    virtual ~ISharedMemory() = default;

    // Initialize Shared Memory for N channels
    // We replace the generic 'size' with specific channel config
    virtual bool Create(const std::string& name, int numChannels, size_t sizePerChannel) = 0;

    // Write a video frame to a specific channel (Thread-Safe via atomics)
    virtual bool WriteFrame(int channelIdx, const uint8_t* data, size_t size, uint64_t timestamp, bool isKey) = 0;

    // Get the raw pointer (mostly for debugging or manual inspection)
    virtual uint8_t* GetBuffer() const = 0;

    virtual void Close() = 0;

    // Static Factory Method: Returns the correct OS version
    static std::shared_ptr<ISharedMemory> CreateInstance();
};
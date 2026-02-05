#pragma once
#include <string>
#include <memory>
#include <cstdint>

// Abstract Interface - pure virtual functions
class ISharedMemory {
public:
    virtual ~ISharedMemory() = default;

    // Allocate (or connect to) the shared memory segment
    virtual bool Create(const std::string& name, size_t size) = 0;

    // Get the raw pointer to write video frames into
    virtual uint8_t* GetBuffer() const = 0;

    // Cleanup resources
    virtual void Close() = 0;

    // Static Factory Method: Returns the correct OS version
    static std::unique_ptr<ISharedMemory> CreateInstance();
};
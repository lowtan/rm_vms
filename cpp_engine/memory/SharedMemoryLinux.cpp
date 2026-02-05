#include "SharedMemory.h"

// Linux-specific headers (Guard these if this file is in a shared build list)
#if defined(__linux__) || defined(__APPLE__)
#include <sys/mman.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <iostream>

class LinuxSharedMemory : public ISharedMemory {
private:
    std::string _name;
    size_t _size = 0;
    int _shm_fd = -1;
    uint8_t* _ptr = nullptr;

public:
    bool Create(const std::string& name, size_t size) override {
        // POSIX SHM requires names to start with "/"
        _name = (name[0] == '/') ? name : "/" + name;
        _size = size;

        // 1. Open Shared Memory Object
        _shm_fd = shm_open(_name.c_str(), O_CREAT | O_RDWR, 0666);
        if (_shm_fd == -1) {
            std::cerr << "shm_open failed" << std::endl;
            return false;
        }

        // 2. Truncate (Set the size of the memory object)
        if (ftruncate(_shm_fd, size) == -1) {
            std::cerr << "ftruncate failed" << std::endl;
            return false;
        }

        // 3. Map to process memory
        void* map = mmap(0, size, PROT_READ | PROT_WRITE, MAP_SHARED, _shm_fd, 0);
        if (map == MAP_FAILED) {
            std::cerr << "mmap failed" << std::endl;
            return false;
        }

        _ptr = static_cast<uint8_t*>(map);
        return true;
    }

    uint8_t* GetBuffer() const override {
        return _ptr;
    }

    void Close() override {
        if (_ptr) {
            munmap(_ptr, _size);
            _ptr = nullptr;
        }
        if (_shm_fd != -1) {
            close(_shm_fd);
            _shm_fd = -1;
        }
        // Optional: shm_unlink(_name.c_str()); 
        // Warning: Unlinking removes the name. Only do this when 
        // you are sure no other process needs to Attach() anymore.
    }

    ~LinuxSharedMemory() {
        Close();
    }
};

// Factory Implementation for Linux
std::unique_ptr<ISharedMemory> ISharedMemory::CreateInstance() {
    return std::make_unique<LinuxSharedMemory>();
}
#endif
#include <queue>
#include <mutex>
#include <condition_variable>
#include <iostream>

template <typename T>

class SafeQueue {
private:
    std::queue<T> queue;
    std::mutex mtx;
    std::condition_variable cv;
    size_t max_size; // Max items before we start dropping frames

public:
    SafeQueue(size_t limit = 1000) : max_size(limit) {}

    // PUSH: Used by the Ingestion Thread
    void push(T item) {
        std::unique_lock<std::mutex> lock(mtx);
        
        // Safety: If queue is full (Disk is too slow?), drop the packet to save RAM.
        if (queue.size() >= max_size) {
            std::cerr << "WARNING: Queue full! Dropping frame." << std::endl;
            // In a real NVR, you might want to free the memory of 'item' here
            // depending on how you manage pointers.
            return; 
        }

        queue.push(item);
        lock.unlock(); // Unlock before notifying to save CPU cycles
        cv.notify_one(); // Wake up the writer thread
    }

    // POP: Used by the Writer Thread
    // This blocks (sleeps) if the queue is empty.
    T pop() {
        std::unique_lock<std::mutex> lock(mtx);
        
        // Wait until queue is NOT empty
        cv.wait(lock, [this] { return !queue.empty(); });

        T item = queue.front();
        queue.pop();
        return item;
    }
};
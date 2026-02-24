package process

import (
	"encoding/binary"
	// "fmt"
	"os"
	"sync/atomic"
	"syscall"
	// "time"
	"unsafe"
)

// CONSTANTS must match C++
const (
	MetadataSize = 64
	HeaderSize   = 64 // RingBufferHeader size
)

// WorkerSHM manages the shared memory connection for one Worker process
type WorkerSHM struct {
	file       *os.File
	data       []byte // Mmap data
	Channels   []*RingBuffer
}

type RingBuffer struct {
	Header    *RingBufferHeader // Points to shared memory
	DataStart []byte            // Slice covering this channel's data area
	Capacity  uint32
}

// C-compatible struct layout representation
type RingBufferHeader struct {
	WriteHead uint32 // Atomic in C++
	ReadTail  uint32 // Atomic in C++
	BufferSize uint32
	StreamID   uint32
	_padding   [48]byte
}

func ConnectSharedMemory(name string, numChannels int, sizePerChannel int) (*WorkerSHM, error) {
	// Open File
	path := "/dev/shm/" + name
    // Wait for C++ to create it, or use a retry loop
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	// Mmap
	stat, _ := f.Stat()
	size := stat.Size()
    
	mmap, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	w := &WorkerSHM{
		file: f,
		data: mmap,
	}

	// 3. Parse Offsets
	offset := 0
	for i := 0; i < numChannels; i++ {
		// Calculate pointer to header
		headerPtr := (*RingBufferHeader)(unsafe.Pointer(&w.data[offset]))
		
		// data starts after header
		dataStartOffset := offset + HeaderSize
		dataEndOffset := dataStartOffset + sizePerChannel

		rb := &RingBuffer{
			Header:    headerPtr,
			DataStart: w.data[dataStartOffset:dataEndOffset],
			Capacity:  uint32(sizePerChannel),
		}
		w.Channels = append(w.Channels, rb)

		offset += HeaderSize + sizePerChannel
	}

	return w, nil
}

// ReadFrame checks if new data is available. 
// This should be called in a loop (goroutine) for each channel.
func (rb *RingBuffer) ReadFrame() ([]byte, uint64, bool) {
	// Atomic Load
	head := atomic.LoadUint32(&rb.Header.WriteHead)
	tail := atomic.LoadUint32(&rb.Header.ReadTail)

	if tail == head {
		return nil, 0, false // Empty
	}

    // Logic to handle Wrap-around detection would go here.
    // In our simplified C++ logic, if Head < Tail, it means Head wrapped.
    // If we are at Tail, check if there is valid data.
    
    // Read Metadata
    // We need to handle the case where Tail wraps to 0 automatically 
    // if C++ wrapped it.
    // Simplification: Assume Tail jumps to 0 if Tail + MetaSize > Capacity
    // (This requires identical logic to C++ writer)

	metaBytes := rb.DataStart[tail : tail+MetadataSize]
	frameSize := binary.LittleEndian.Uint32(metaBytes[0:4])
	timestamp := binary.LittleEndian.Uint64(metaBytes[8:16])

	// Calculate data location
	start := tail + MetadataSize
	end := start + frameSize

	// Copy data out to Go memory (Safe)
    // For true zero-copy, you'd pass the slice slice := rb.DataStart[start:end]
    // BUT you must process it before updating ReadTail.
	frameData := make([]byte, frameSize)
	copy(frameData, rb.DataStart[start:end])

	// Update Tail (Commit Read)
	newTail := end
    
    // Handle the C++ wrap logic:
    // If the Writer wrapped, we must also wrap if we hit the end constraint.
    // In production, you usually use a 'magic byte' or specific flag 
    // in the header to indicate "End of Buffer, go to 0".
    
	atomic.StoreUint32(&rb.Header.ReadTail, newTail)

	return frameData, timestamp, true
}

func (w *WorkerSHM) Close() {
	syscall.Munmap(w.data)
	w.file.Close()
}
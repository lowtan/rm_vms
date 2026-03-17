package shm

import (
	// "encoding/binary"
	// "log"
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
	MagicNumber  uint32 = 0xFFAABBCC
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

	// Parse Offsets
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

func (rb *RingBuffer) ReadFrame() (FrameMetadata, []byte, bool) {
	head := atomic.LoadUint32(&rb.Header.WriteHead)
	tail := atomic.LoadUint32(&rb.Header.ReadTail)

	if tail == head {
		return FrameMetadata{}, nil, false // Buffer is empty
	}

	// Edge Case: Tail is so close to the end that even the 64-byte metadata won't fit.
	// C++ definitely wrapped here.
	if tail+MetadataSize > rb.Capacity {
		tail = 0
	}

	// Read Metadata
	meta := FrameMetadata{}
	metaBytes := rb.DataStart[tail : tail+MetadataSize]
	magic := meta.GetMagic(metaBytes)

	// Wrap-Around Detection
	if magic != MagicNumber {
		// C++ skipped this section because the frame didn't fit. 
		// Wrap the reader to 0.
		tail = 0

		// Re-read metadata at index 0
		metaBytes = rb.DataStart[tail : tail+MetadataSize]
		magic = meta.GetMagic(metaBytes)

		// If it's STILL wrong, the reader has fallen completely out of sync 
		// (e.g., C++ overwrote the data before Go could read it).
		if magic != MagicNumber {
			// Emergency Recovery: Catch up to the writer's head to drop corrupted state
			atomic.StoreUint32(&rb.Header.ReadTail, head)
			return FrameMetadata{}, nil, false
		}
	}

	meta.LoadFrom(metaBytes)
	// Extract Metadata
	// Note: frameSize is now at offset 4:8 because magic takes 0:4
	// frameSize := binary.LittleEndian.Uint32(metaBytes[4:8])
	// timestamp := binary.LittleEndian.Uint64(metaBytes[8:16])
	// isKeyFrame := metaBytes[16] !=0 // Available if you need it later
	// mediaType := metaBytes[17]

	start := tail + MetadataSize
	end := start + meta.FrameSize

	// Ultimate Panic Fail-Safe
	// Even with the magic number, if memory corruption occurs, prevent a crash.
	if end > rb.Capacity {
		atomic.StoreUint32(&rb.Header.ReadTail, head) // Drop and recover
		return FrameMetadata{}, nil, false
	}

	// Copy Data safely
	frameData := make([]byte, meta.FrameSize)
	copy(frameData, rb.DataStart[start:end])

	// Commit Read
	atomic.StoreUint32(&rb.Header.ReadTail, end)

	return meta, frameData, true
}


func (w *WorkerSHM) Close() {
	syscall.Munmap(w.data)
	w.file.Close()
}
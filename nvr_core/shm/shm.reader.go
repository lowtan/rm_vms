package shm

import (
	// "context"
	"fmt"
	"log"
	"nvr_core/stream"
	"os"
	"sync/atomic"
	"time"
)

type ReaderSHM struct {
	workerName     string
	worker         *WorkerSHM
	camChannels    map[int]int
	channelHub     map[int]*stream.Hub
	channelStopper map[int]*atomic.Bool
}

// StartStreamReader connects to a Worker's SHM and starts reading its channels.
// workerName: e.g., "worker1" (Matches the C++ name creation logic)
// numChannels: 16 (Based on your sharding strategy)
// bufferSize: 1500000 (The 1.5MB calculated earlier)
func StartStreamReader(workerName string, numChannels int, bufferSize int) (*ReaderSHM) {
	shmName := "shm" + workerName
	var shm *WorkerSHM
	var err error

	// Retry loop to wait for C++ to create the file
	for i := 0; i < 50; i++ {
		shm, err = ConnectSharedMemory(shmName, numChannels, bufferSize)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond) // Wait 100ms and try again
	}

	if err != nil {
		log.Printf("[Go Manager] Failed to connect to SHM %s after retries: %v\n", shmName, err)
		return nil
	}
	
	log.Printf("[Go Manager] Successfully connected to %s. Spawning the reader...\n", shmName)

	reader := ReaderSHM{
		workerName: workerName,
		worker: shm,
		camChannels: make(map[int]int),
		channelHub: make(map[int]*stream.Hub),
		channelStopper: make(map[int]*atomic.Bool),
	}

	// TODO: In production, you would attach a context.Context here 
	// to handle graceful shutdown and call shm.Close() when the worker dies.

	// context.Context

	log.Printf("[Go Manager] %s reader spawned\n", shmName)

	return &reader;
}

func (r *ReaderSHM) FilePath() string {
	if r.worker != nil && r.worker.file != nil {
		return r.worker.file.Name() // Returns exactly what was passed to os.OpenFile
	}
	return ""
}

func (r *ReaderSHM) StartChannel(channelID int, existingHub *stream.Hub) *stream.Hub {

	rb := r.worker.Channels[channelID]
	if(rb == nil) {

		log.Printf("[startChannel][%d] no ring buffer!\n", channelID);
		return nil;
	}

	hub := existingHub
	if(hub == nil) {

		log.Printf("[startChannel][%d] streamer starting!\n", channelID);
		hub = stream.NewHub()
		go hub.Run()

	}

	r.channelHub[channelID] = hub
	stopper := &atomic.Bool{}
	stopper.Store(false)
	r.channelStopper[channelID] = stopper

	go r.readChannelLoop(stopper, channelID, rb, hub)

	return hub

}

func (r *ReaderSHM) stopChannel(channelID int) {
	stopper := r.channelStopper[channelID]
	if stopper == nil {
		return
	}
	stopper.Store(true);
}

// readChannelLoop continuously polls a specific RingBuffer for new frames
func (r *ReaderSHM) readChannelLoop(stop *atomic.Bool, channelID int, rb *RingBuffer, bc *stream.Hub) {
	log.Printf("[Go][shm][readChannelLoop] Started reading loop for %s Channel %d\n", r.workerName, channelID)

	for {

		if(stop.Load()) {
			break;
		}

		// Attempt to read a frame
		// frameData, timestamp, ok := rb.ReadFrame()
		frameData, _, isKey, mediaType, ok := rb.ReadFrame()

		if !ok {
			// Buffer is empty.
			// Sleep for 1 millisecond to prevent the Goroutine from pegging the CPU to 100%.
			// At 30 FPS, a frame arrives roughly every ~33ms, so 1ms polling is highly responsive.
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// --- ZERO-COPY READ SUCCESS ---
		// frameData now holds the raw video bytes (e.g., H.264 NAL units).
		// You can route this data to disk, a WebSocket, or a WebRTC pipeline.

		f := stream.StreamPacket { IsKeyFrame: isKey, Payload: frameData, MediaType: mediaType }

		bc.Broadcast <- f
		// fileDumpTest(frameData, r.workerName, channelID)

	}

}

// Close safely stops all polling goroutines and unmaps the shared memory
func (r *ReaderSHM) Close() {
	log.Printf("[shm.reader] Closing SHM Reader for %s\n", r.workerName)
	
	// 1. Signal all readChannelLoop goroutines to exit
	for _, stopper := range r.channelStopper {
		stopper.Store(true)
	}

	// 2. Unmap the memory to prevent OS RAM leaks
	if r.worker != nil {
		r.worker.Close()
	}
}

func fileDumpTest(frame []byte, worker string, channel int) {

	f, err := os.OpenFile(fmt.Sprintf("/app/recordings/debug_worker%v_cam_%d.264", worker, channel), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(frame); err != nil {
		log.Println("Write error:", err)
	}

}
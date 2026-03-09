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
	
	log.Printf("[Go Manager] Successfully connected to %s. Spawning readers...\n", shmName)

	reader := ReaderSHM{
		workerName: workerName,
		worker: shm,
		camChannels: make(map[int]int),
		channelHub: make(map[int]*stream.Hub),
		channelStopper: make(map[int]*atomic.Bool),
	}

	// XXXXXXXXXXXXXX
	// We are not going to read all the channels
	// boot them up separately after we get camera starting
	// status.
	// 
	// Spawn a Goroutine for each camera channel
	// for i := 0; i < numChannels; i++ {
	// 	go readChannelLoop(shmName, i, shm.Channels[i])
	// }

	// Note: In production, you would attach a context.Context here 
	// to handle graceful shutdown and call shm.Close() when the worker dies.

	// context.Context

	return &reader;
}



func (r *ReaderSHM) StartChannel(channelID int) *stream.Hub {

	rb := r.worker.Channels[channelID]
	if(rb == nil) {

		log.Printf("[startChannel][%d] no ring buffer!\n", channelID);
		return nil;

	} else {

		hub := r.channelHub[channelID]
		if(hub == nil) {

			log.Printf("[startChannel][%d] streamer starting!\n", channelID);
			hub = stream.NewHub()
			r.channelHub[channelID] = hub

			go hub.Run()

			stopper := &atomic.Bool{}
			stopper.Store(false)
			r.channelStopper[channelID] = stopper
			go r.readChannelLoop(stopper, channelID, rb, hub)

		} else {

			log.Printf("[startChannel][%d] streamer is already running!\n%v", channelID, hub);

		}

		return hub

	}

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
		frameData, _, isKey, ok := rb.ReadFrame()

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

		f := stream.VideoFrame { IsKeyFrame: isKey, Payload: frameData }

		bc.Broadcast <- f
		fileDumpTest(frameData, r.workerName, channelID)

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
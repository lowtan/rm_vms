package process

import (
	// "context"
	"fmt"
	"log"
	"time"
)

// StartStreamReader connects to a Worker's SHM and starts reading its channels.
// workerName: e.g., "worker1" (Matches the C++ name creation logic)
// numChannels: 16 (Based on your sharding strategy)
// bufferSize: 1500000 (The 1.5MB calculated earlier)
func StartStreamReader(workerName string, numChannels int, bufferSize int) {
	shmName := "shm" + workerName
	var shm *WorkerSHM
	var err error

	// 1. Retry loop to wait for C++ to create the file
	for i := 0; i < 50; i++ {
		shm, err = ConnectSharedMemory(shmName, numChannels, bufferSize)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond) // Wait 100ms and try again
	}

	if err != nil {
		log.Printf("[Go Manager] Failed to connect to SHM %s after retries: %v\n", shmName, err)
		return
	}
	
	log.Printf("[Go Manager] Successfully connected to %s. Spawning readers...\n", shmName)

	// 2. Spawn a Goroutine for each camera channel
	for i := 0; i < numChannels; i++ {
		go readChannelLoop(shmName, i, shm.Channels[i])
	}

	// Note: In production, you would attach a context.Context here 
	// to handle graceful shutdown and call shm.Close() when the worker dies.

	// context.Context

}

// readChannelLoop continuously polls a specific RingBuffer for new frames
func readChannelLoop(workerName string, channelID int, rb *RingBuffer) {
	log.Printf("[Go Manager] Started reading loop for %s Channel %d\n", workerName, channelID)

	for {
		// Attempt to read a frame
		frameData, timestamp, ok := rb.ReadFrame()

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

		// For debugging, print occasionally (e.g., every 100th frame)
		if timestamp%1000 == 0 {
			fmt.Printf("[%s-Ch%d] Received Frame: %d bytes, TS: %d\n", 
				workerName, channelID, len(frameData), timestamp)
		}
	}
}
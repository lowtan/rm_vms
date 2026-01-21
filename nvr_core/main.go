package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"
)

// Matches the JSON sent from C++
type WorkerResponse struct {
	Status string `json:"status"`
	CamID  int    `json:"cam"`
}

func main() {

	// Load Configuration
	fmt.Println("[Go Manager] Loading config...")

	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("[Go Manager] Config Loaded. Storage: %s, Cameras: %d\n", 
			cfg.Server.StoragePath, len(cfg.Cameras))


	// 1. Spawn the C++ Binary
	cmd := exec.Command("./nvr_worker")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start worker:", err)
	}
	fmt.Println("[Go Manager] C++ Worker started.")

	// 2. Listen for STDOUT (JSON Responses)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()
			var resp WorkerResponse
			// If it's valid JSON, print prettily
			if err := json.Unmarshal([]byte(text), &resp); err == nil {
				fmt.Printf("[Go Manager] Received Update -> Cam %d Status: %s\n", resp.CamID, resp.Status)
			} else {
				fmt.Println("[Go Manager] Raw:", text)
			}
		}
	}()

	// 3. Listen for STDERR (Logs)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// Print C++ logs in Red color
			fmt.Printf("\033[31m%s\033[0m\n", scanner.Text())
		}
	}()


	// Loop through cameras and start the ones that are enabled
	for _, cam := range cfg.Cameras {

		fmt.Printf("[Go Manager] Starting Camera %d (%s)...\n", cam.ID, cam.Name)
		
		// Construct the command string: "START <ID> <URL>"
		cmdStr := fmt.Sprintf("START %d %s\n", cam.ID, cam.URL)

		// Send to C++ Worker via StdinPipe
		io.WriteString(stdin, cmdStr)

	}

	// 4. Send a Test Command
	// Note: using a public test RTSP stream (if available) or a fake one
	// time.Sleep(1 * time.Second)
	// command := "START 1 rtsp://host.docker.internal:8554/mystream"
	// fmt.Println("[Go Manager] Sending:", command)
	// io.WriteString(stdin, command+"\n")

	// 5. Keep alive for 5 seconds then exit
	time.Sleep(60 * time.Second)
	io.WriteString(stdin, "EXIT\n")
	cmd.Wait()
	fmt.Println("[Go Manager] Exiting.")
}
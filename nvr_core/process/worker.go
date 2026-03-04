package process

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const LOGSEP = "==============================================\n"

// Matches the JSON sent from C++
type WorkerResponse struct {
    Status string `json:"status"`
    CamID  int    `json:"cam"`
}

type Camera struct {
    ID   int
    url  string
}

// Worker represents a single C++ subprocess
type Worker struct {
    ID         int
    BinaryPath string
    Cmd        *exec.Cmd
    Stdin      io.WriteCloser
    Stdout      io.ReadCloser
    Stderr      io.ReadCloser
    cameras    map[int]Camera
    mu         sync.Mutex // Protects concurrent writes to Stdin
}

// NewWorker creates a struct but doesn't start the process yet
func NewWorker(id int, binaryPath string) *Worker {
    return &Worker{
        ID:         id,
        BinaryPath: binaryPath,
        cameras: make(map[int]Camera),
    }
}

func (w *Worker) handleStoppedStream(resp WorkerResponse) {

    time.Sleep(8 * time.Second)

    fmt.Println(LOGSEP + "[Go][Worker] restarting cam:", resp.CamID)
    w.RestartCam(resp.CamID)

}

func (w *Worker) startSHMReader(resp WorkerResponse) {

    // Launch the reader in the background so it doesn't block the command loop
    go StartStreamReader(strconv.Itoa(w.ID), 10, 3145728)

}

// Start launches the C++ binary and sets up pipes
// This code will setup pipes and send WorkerID to
// cpp program, and should not be called twice.
func (w *Worker) Start(ctx context.Context) error {

    w.Cmd = exec.CommandContext(ctx, w.BinaryPath)

    // Configure Graceful Termination (Crucial for /dev/shm cleanup)
    // By default, CommandContext sends a brutal SIGKILL when the context cancels.
    // We override this to send a SIGTERM first, giving your C++ destructors a chance to run.
    w.Cmd.Cancel = func() error {
        // log.Printf("[Worker %s] Sending SIGTERM for graceful shutdown...", workerID)
        return w.Cmd.Process.Signal(syscall.SIGTERM)
    }
    
    // Give the C++ worker 3 seconds to unmap memory and exit smoothly. 
    // If it hangs, Go will follow up with a SIGKILL to forcefully terminate it.
    w.Cmd.WaitDelay = 3 * time.Second

    // Setup Stdin Pipe
    if err := w.ConfigureCMDPipes(); err != nil {
        return err;
    }

    if err := w.Cmd.Start(); err != nil {
        return fmt.Errorf("worker %d start failed: %v", w.ID, err)
    }

    w.connectCMDIPC()
    w.SendWorkerID()

    // Launch the monitor in the background. 
    // This goroutine will sleep quietly until the C++ process dies.
    go w.monitorProcess(ctx)

    return nil
}

// SendCommand is a thread-safe way to write to the worker
func (w *Worker) SendCommand(cmd string) error {
    w.mu.Lock()
    defer w.mu.Unlock()

    if w.Stdin == nil {
        return fmt.Errorf("worker %d is not running", w.ID)
    }

    // Add newline because C++ uses std::getline
    _, err := io.WriteString(w.Stdin, cmd+"\n")
    return err
}

func (w *Worker) SendWorkerID() error {

    command := fmt.Sprintf("WORKER %d", w.ID)
    return w.SendCommand(command)

}


func (w *Worker) AssignCam(cam Camera) error {

    w.mu.Lock()
    w.cameras[cam.ID] = cam
    w.mu.Unlock()

    return w.StartCam(cam)

}

func (w *Worker) StartCam(cam Camera) error {

    command := fmt.Sprintf("START %d %s", cam.ID, cam.url)
    return w.SendCommand(command)

}

func (w *Worker) RestartCam(camID int) error {

    w.mu.Lock()
    cam := w.cameras[camID]
    w.mu.Unlock()

    return w.StartCam(cam)

}


// Stop cleanly closes the pipe and waits
func (w *Worker) Stop() {
    if w.Stdin != nil {
        w.Stdin.Close() // Sends EOF to C++
    }
    if w.Stdout != nil {
        w.Stdout.Close() // Sends EOF to C++
    }
    if w.Stderr != nil {
        w.Stderr.Close() // Sends EOF to C++
    }
    if w.Cmd != nil {
        w.Cmd.Wait()
    }
}


/**
 * ======================================================
 * Logging and IPC Response
 * ======================================================
 */

func (w *Worker) handleCMDResponse(resp WorkerResponse) {
    if(resp.Status == "stopped") {

        w.handleStoppedStream(resp)

    } else if(resp.Status == "starting") {

        w.startSHMReader(resp)

    }
}

func (w *Worker) connectCMDIPC() {

    // Listen for STDOUT (JSON Responses)
    go func() {
        scanner := bufio.NewScanner(w.Stdout)
        for scanner.Scan() {
            text := scanner.Text()
            var resp WorkerResponse
            // If it's valid JSON, print prettily
            if err := json.Unmarshal([]byte(text), &resp); err == nil {
                fmt.Printf("[Go][Worker] Received Update -> Cam %d Status: %s\n", resp.CamID, resp.Status)
                w.handleCMDResponse(resp);
            } else {
                fmt.Printf("\033[38;2;0;200;0m%s\033[0m\n", text)
            }
        }
    }()

    // Listen for STDERR (Logs)
    go func() {
        scanner := bufio.NewScanner(w.Stderr)
        for scanner.Scan() {
            // Print C++ logs in Red color
            fmt.Printf("\033[31m%s\033[0m\n", scanner.Text())
        }
    }()

}

// Setup ALL Pipes
// This must be done BEFORE starting the command
func (w *Worker) ConfigureCMDPipes() error {
    stdin, err := w.Cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdin error: %v", w.ID, err)
    }
    w.Stdin = stdin

    stdout, err := w.Cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdout error: %v", w.ID, err)
    }
    w.Stdout = stdout

    stderr, err := w.Cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("worker %d stderr error: %v", w.ID, err)
    }
    w.Stderr = stderr

    return nil
}


/**
 * ======================================================
 * Monitor and Recover
 * ======================================================
 */
// monitorProcess waits for the C++ binary to exit and triggers a restart if it was unexpected.
func (w *Worker) monitorProcess(ctx context.Context) {

    // Wait blocks until the C++ process exits.
    err := w.Cmd.Wait()

    // Check if this was an intentional shutdown from the Go Manager
    if ctx.Err() != nil {
        fmt.Printf("[Go][Worker %d] Shut down gracefully.\n", w.ID)
        return
    }

    // If we reach here, the C++ process died unexpectedly!
    if err != nil {
        // Log in red for visibility
        fmt.Printf("\033[31m[Go][Worker %d] CRASH DETECTED: %v\033[0m\n", w.ID, err)
    } else {
        fmt.Printf("\033[31m[Go][Worker %d] Exited unexpectedly with code 0.\033[0m\n", w.ID)
    }

    w.recoverWorker(ctx)

}

// recoverWorker attempts to restart the C++ binary and restore its previous state
func (w *Worker) recoverWorker(ctx context.Context) {
    fmt.Printf("[Go][Worker %d] Attempting to restart in 3 seconds...\n", w.ID)
    
    // Add a small backoff delay. If the C++ worker is crashing instantly on startup
    // (e.g., due to a bad config), this prevents an infinite CPU-burning crash loop.
    time.Sleep(3 * time.Second)

    // --- THE CLEANUP STRATEGY ---
    // TODO: we shall deal with SHM data corruption issue later
    // w.mu.Lock()
    // for _, cam := range w.cameras {
    //     // Define the path to the shared memory block for this specific camera
    //     shmPath := fmt.Sprintf("/dev/shm/nvr_buffer_cam_%d", cam.ID)
        
    //     // Delete the corrupted memory block. 
    //     // We ignore os.IsNotExist errors in case the C++ worker never created it.
    //     if err := os.Remove(shmPath); err != nil && !os.IsNotExist(err) {
    //         fmt.Printf("\033[33m[Go][Worker %d] Warning: failed to unlink SHM for cam %d: %v\033[0m\n", w.ID, cam.ID, err)
    //     } else {
    //         fmt.Printf("[Go][Worker %d] Cleared shared memory for cam %d\n", w.ID, cam.ID)
    //     }
    // }
    // w.mu.Unlock()
    // ----------------------------

    // Restart the process
    if err := w.Start(ctx); err != nil {
        fmt.Printf("[Go][Worker %d] Failed to restart: %v\n", w.ID, err)
        // In a production app, you might want to retry with an exponential backoff here
        return
    }

    // Restore State: Re-assign all cameras that this worker was managing
    w.mu.Lock()
    defer w.mu.Unlock()
    for _, cam := range w.cameras {
        fmt.Printf("[Go][Worker %d] Restoring stream for Cam %d...\n", w.ID, cam.ID)
        w.StartCam(cam)
    }
}
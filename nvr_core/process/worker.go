package process

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
	"os/exec"
	"sync"
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

func handleStoppedStream(w *Worker, resp WorkerResponse) {
    if(resp.Status == "stopped") {

        time.Sleep(8 * time.Second)

        fmt.Println(LOGSEP + "[Go][Worker] restarting cam:", resp.CamID)
        w.RestartCam(resp.CamID)

    }
}

// Start launches the C++ binary and sets up pipes
// This code will setup pipes and send WorkerID to
// cpp program, and should not be called twice.
func (w *Worker) Start() error {
    w.Cmd = exec.Command(w.BinaryPath)

    // Setup Stdin Pipe
    stdin, err := w.Cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdin error: %v", w.ID, err)
    }
    w.Stdin = stdin

    // Redirect Stdout/Stderr to parent for now (Logs)
    // w.Cmd.Stdout = os.Stdout
    // w.Cmd.Stderr = os.Stderr

    stdoutp, _ := w.Cmd.StdoutPipe()
    stderrp, _ := w.Cmd.StderrPipe()

    if err := w.Cmd.Start(); err != nil {
        return fmt.Errorf("worker %d start failed: %v", w.ID, err)
    }

    // fmt.Printf("[Go][Worker] start stdin/err scanner, %v %v\n", stdoutp, stderrp)

    // Listen for STDOUT (JSON Responses)
    go func() {
        scanner := bufio.NewScanner(stdoutp)
        for scanner.Scan() {
            text := scanner.Text()
            var resp WorkerResponse
            // If it's valid JSON, print prettily
            if err := json.Unmarshal([]byte(text), &resp); err == nil {
                fmt.Printf("[Go][Worker] Received Update -> Cam %d Status: %s\n", resp.CamID, resp.Status)
                handleStoppedStream(w, resp);
            } else {
                fmt.Printf("\033[38;2;0;200;0m%s\033[0m\n", text)
            }
        }
    }()

    // Listen for STDERR (Logs)
    go func() {
        scanner := bufio.NewScanner(stderrp)
        for scanner.Scan() {
            // Print C++ logs in Red color
            fmt.Printf("\033[31m%s\033[0m\n", scanner.Text())
        }
    }()

    w.SendWorkerID()

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

    w.cameras[cam.ID] = cam;
    return w.StartCam(cam)

}

func (w *Worker) StartCam(cam Camera) error {

    command := fmt.Sprintf("START %d %s", cam.ID, cam.url)
    return w.SendCommand(command)

}

func (w *Worker) RestartCam(camID int) error {

    cam := w.cameras[camID]
    return w.StartCam(cam)

}


// Stop cleanly closes the pipe and waits
func (w *Worker) Stop() {
    if w.Stdin != nil {
        w.Stdin.Close() // Sends EOF to C++
    }
    if w.Cmd != nil {
        w.Cmd.Wait()
    }
}
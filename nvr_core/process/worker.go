package process

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "sync"
)

// Worker represents a single C++ subprocess
type Worker struct {
    ID         int
    BinaryPath string
    Cmd        *exec.Cmd
    Stdin      io.WriteCloser
    mu         sync.Mutex // Protects concurrent writes to Stdin
}

// NewWorker creates a struct but doesn't start the process yet
func NewWorker(id int, binaryPath string) *Worker {
    return &Worker{
        ID:         id,
        BinaryPath: binaryPath,
    }
}

// Start launches the C++ binary and sets up pipes
func (w *Worker) Start() error {
    w.Cmd = exec.Command(w.BinaryPath)

    // Setup Stdin Pipe
    stdin, err := w.Cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdin error: %v", w.ID, err)
    }
    w.Stdin = stdin

    // Redirect Stdout/Stderr to parent for now (Logs)
    w.Cmd.Stdout = os.Stdout
    w.Cmd.Stderr = os.Stderr

    if err := w.Cmd.Start(); err != nil {
        return fmt.Errorf("worker %d start failed: %v", w.ID, err)
    }

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

// Stop cleanly closes the pipe and waits
func (w *Worker) Stop() {
    if w.Stdin != nil {
        w.Stdin.Close() // Sends EOF to C++
    }
    if w.Cmd != nil {
        w.Cmd.Wait()
    }
}
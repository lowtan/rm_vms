package process

import (
	"fmt"
	"log"
)

// Manager handles the pool of workers
type Manager struct {
	workers    []*Worker
	binaryPath string
}

// NewManager initializes the pool (e.g., count=4)
func NewManager(count int, binaryPath string) *Manager {
	mgr := &Manager{
		workers:    make([]*Worker, count),
		binaryPath: binaryPath,
	}

	// Initialize workers
	for i := 0; i < count; i++ {
		mgr.workers[i] = NewWorker(i, binaryPath)
	}

	return mgr
}

// StartAll launches all worker processes
func (m *Manager) StartAll() error {
	for _, w := range m.workers {
		fmt.Printf("[Process Manager] Starting Worker %d...\n", w.ID)
		if err := w.Start(); err != nil {
			return err
		}
	}
	return nil
}

// StopAll shuts them down
func (m *Manager) StopAll() {
	for _, w := range m.workers {
		w.Stop()
	}
}

// AssignCamera routes a camera to the correct worker (Sharding Logic)
func (m *Manager) AssignCamera(camID int, url string) error {
	if len(m.workers) == 0 {
		return fmt.Errorf("no workers available")
	}

	// SHARDING ALGORITHM: Round Robin using Modulus
	workerIndex := camID % len(m.workers)
	targetWorker := m.workers[workerIndex]

	log.Printf("[Process Manager] Routing Cam %d -> Worker %d\n", camID, workerIndex)

	return targetWorker.AssignCam(Camera{ camID, url })
}
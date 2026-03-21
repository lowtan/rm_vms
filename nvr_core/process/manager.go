package process

import (
	"context"
	"fmt"
	"log"
	"nvr_core/utils"
	"nvr_core/db/ingest"
)



// Manager handles the pool of workers
type Manager struct {
	cfg        *utils.Config
	ctx        context.Context
	workers    []*Worker
	binaryPath string
	camWorker  map[int]int
	ingester   *ingest.BatchIngester
}

// NewManager initializes the pool (e.g., count=4)
func NewManager(ctx context.Context, cfg *utils.Config, count int, binaryPath string, ingester *ingest.BatchIngester) *Manager {
	mgr := &Manager{
		ctx: ctx,
		cfg: cfg,
		workers:    make([]*Worker, count),
		binaryPath: binaryPath,
		camWorker: make(map[int]int),
		ingester:   ingester,
	}

	// Initialize workers
	for i := 0; i < count; i++ {
		mgr.workers[i] = NewWorker(i, binaryPath, ingester)
	}

	return mgr
}

// StartAll launches all worker processes
func (m *Manager) StartAll() error {
	for _, w := range m.workers {
		fmt.Printf("[Process Manager] Starting Worker %d...\n", w.ID)
		if err := w.Start(m.ctx); err != nil {
			return err
		}
	}
	return nil
}

// StopAll shuts them down
func (m *Manager) StopAll() {
	log.Printf("[Manager][StopAll]")
	for _, w := range m.workers {
		w.Stop()
	}
}

func (m *Manager) GetWorkers() []*Worker {
	return m.workers
}

// AssignCamera routes a camera to the correct worker (Sharding Logic)
func (m *Manager) AssignCamera(camID int, url string) error {
	if len(m.workers) == 0 {
		return fmt.Errorf("no workers available")
	}

	// SHARDING ALGORITHM: Round Robin using Modulus
	workerIndex := camID % len(m.workers)
	targetWorker := m.workers[workerIndex]

	m.camWorker[camID] = workerIndex;

	log.Printf("[Manager][AssignCamera] Routing Cam %d -> Worker %d\n", camID, workerIndex)

	return targetWorker.AssignCam(NewCamera(camID, url))
}

func (m *Manager) CameraWorker(camID int) *Worker {

	log.Printf("[Manager][CameraWorker] Cam %d list(%d)\n", camID, len(m.camWorker))

	index := m.camWorker[camID];

	log.Printf("[Manager][CameraWorker] Cam %d -> Worker %d\n", camID, index)

	return m.workers[index];
}
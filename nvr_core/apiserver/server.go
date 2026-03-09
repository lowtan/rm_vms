package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"nvr_core/process"
	"nvr_core/utils"
)

// Camera represents the configuration of an RTSP stream
type Camera struct {
	ID       string `json:"id"`
	RTSPUrl  string `json:"rtsp_url"`
	WorkerID string `json:"worker_id"`
	Status   string `json:"status"`
}

// NVRState uses sync.Map for highly concurrent, lock-free (mostly) reads/writes
type NVRState struct {
	Cameras sync.Map
}

func NewNVRState() *NVRState {
	return &NVRState{}
}

type APIServer struct {
	CFG   *utils.Config
	State *NVRState
	PM    *process.Manager
	// SHub *stream.Hub
}

func Initiate(ctx context.Context, cfg *utils.Config, pm *process.Manager) {

	log.Println("Starting API server")

	state := NewNVRState()

	api := &APIServer{State: state, CFG: cfg, PM: pm}

	mux := http.NewServeMux()

	// mux.HandleFunc("GET /ws/stream/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	// stream.ServeWs(hub, w, r)
	// })

	mux.HandleFunc("GET /ws/stream/{id}", api.GetStream)

	mux.HandleFunc("GET /health", api.GetHealth)

	mux.HandleFunc("GET /api/cameras", api.GetCameras)
	mux.HandleFunc("POST /api/cameras", api.AddCamera)

	addr := fmt.Sprintf(":%d", cfg.Server.Port);

	// Production-grade server configuration
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,  // Max time to read request headers/body
		WriteTimeout: 10 * time.Second, // Max time to write response
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}


	// Start the server in a background goroutine
	go func() {
		log.Printf("[API] Server listening on %s", addr)
		// ErrServerClosed is expected when we call srv.Shutdown()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] Server critically failed: %v", err)
		}
	}()

	// Block this function until the parent context is canceled (SIGTERM/SIGINT)
	<-ctx.Done()

	log.Println("[API] Shutdown signal received. Finishing active requests...")


	// Create a secondary timeout context specifically for the shutdown process.
	// This ensures a malicious or ultra-slow client can't hold the server open forever.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute the graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[API] Server forced to shutdown due to timeout: %v", err)
	} else {
		log.Println("[API] Server gracefully stopped.")
	}

}
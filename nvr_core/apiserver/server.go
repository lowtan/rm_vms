package apiserver

import (
	"context"
	// "database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	// "nvr_core/db/repository"
	"nvr_core/process"
	"nvr_core/utils"
	"nvr_core/service"
)


// NVRState uses sync.Map for highly concurrent, lock-free (mostly) reads/writes
type NVRState struct {
	Cameras sync.Map
}

func NewNVRState() *NVRState {
	return &NVRState{}
}

type APIServer struct {
	Context context.Context
	CFG   *utils.Config
	State *NVRState
	PM    *process.Manager
	Services *service.Services
}

func Initiate(ctx context.Context, cfg *utils.Config, pm *process.Manager, svcs *service.Services) {

	// segRepo := repository.NewSegmentRepository(dbConn)
	// dbH := NewDebugHandler(ctx, dbConn, segRepo)

	log.Println("Initializing API server")

	state := NewNVRState()

	api := &APIServer{
		Context: ctx,
		State: state,
		CFG: cfg,
		PM: pm,
		Services: svcs,
	}

	mux := http.NewServeMux()

	// Debug Info
	mux.HandleFunc("GET /debug/db", api.GetDebugInfo)

	// Get camera stream
	mux.HandleFunc("GET /ws/stream/{id}", api.GetStream)

	mux.HandleFunc("GET /health", api.GetHealth)

	mux.HandleFunc("GET /api/cameras", api.GetCameras)
	mux.HandleFunc("POST /api/cameras", api.AddCamera)

	// Timeline and Playback
	mux.HandleFunc("POST /api/cameras/{cam_id}/timeline/{start}/{end}", api.GetTimeline)
	mux.HandleFunc("GET /api/cameras/{cam_id}/play", api.HandlePlayVideo)


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
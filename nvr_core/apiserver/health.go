package apiserver

import (
	// "context"
	"encoding/json"
	"log"
	"net/http"
	// "sync"
	// "time"
)

// A quick health check path

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("ok"); err != nil {
		log.Printf("Error checking health: %v", err)
	}

	// TODO: return different HTTP Code for more sound health check

}

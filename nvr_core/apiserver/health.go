package apiserver

import (
	"log"
	"net/http"
)

// A quick health check path

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {

	if err := RespondJSON(w, "ok"); err != nil {
		log.Printf("Error checking health: %v", err)
	}

	// TODO: return different HTTP Code for more sound health check

}

package apiserver

import (
	// "context"
	"encoding/json"
	"log"
	"net/http"
	// "sync"
	// "time"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetCameras(w http.ResponseWriter, r *http.Request) {
	var camList []Camera

	// Distribute Cameras
	for _, cam := range s.CFG.Cameras {
		if cam.Enabled {
			camList = append(camList, Camera{
				ID: cam.Name,
				RTSPUrl: cam.URL,
				Status: "",
			})
		}
	}


	// s.State.Cameras.Range(func(key, value any) bool {
	// 	if cam, ok := value.(Camera); ok {
	// 		camList = append(camList, cam)
	// 	}
	// 	return true // continue iteration
	// })

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(camList); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

// AddCamera stores the camera and triggers the IPC pipeline
func (s *APIServer) AddCamera(w http.ResponseWriter, r *http.Request) {
	var newCam Camera
	
	// Limit request body size to prevent memory exhaustion attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1024*10) // 10KB max
	if err := json.NewDecoder(r.Body).Decode(&newCam); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	newCam.Status = "initializing"
	s.State.Cameras.Store(newCam.ID, newCam)

	// TODO: Send JSON payload via Stdin to the target C++ Worker Subprocess

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newCam); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

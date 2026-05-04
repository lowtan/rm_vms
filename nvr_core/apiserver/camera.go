package apiserver

import (
	// "context"
	"encoding/json"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/process"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetCameras(w http.ResponseWriter, r *http.Request) {
	var camList []*process.Camera

	workers := s.PM.GetWorkers()

	log.Printf("[GetCameras] workers(%d)\n", len(workers))

	for _, worker := range workers {

		cams := worker.GetCameras()
		log.Printf("[GetCameras] cams (%d)\n", len(cams))
		for _, cam := range cams {
			// cam.rtsp = ""
			camList = append(camList, cam)
		}

	}

	log.Printf("[GetCameras] camList(%d)\n", len(camList))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(camList); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

func (s *APIServer) GetDBCameras(w http.ResponseWriter, r *http.Request) {

	camList, err := s.Services.Camera.GetAll(r.Context())

	if(err != nil) {
		log.Printf("Error getting database camera list: %v", err)
		http.Error(w, "failed to get db cameras.", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetCameras] camList(%d)\n", len(camList))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(camList); err != nil {
		log.Printf("Error encoding camera list: %v", err)
		// Connection likely dropped; no need to write http.Error
	}
}

/**
JSON Payload {
	name: string
	manufacturer: string
	model: string
	serial_number: string

	ip_address: string
	http_port: int
	type: string

	username: string
	password: string

	stream_url: string
	sub_stream_url: string
		
	retention_gb_limit: int
	is_active: bool
}
 */
// AddCamera stores the camera and triggers the IPC pipeline
func (s *APIServer) AddCamera(w http.ResponseWriter, r *http.Request) {

	var newCamera dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&newCamera); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.Services.Camera.AddCamera(ctx, newCamera.MapToDBCamera()); err != nil {
		log.Printf("Failed to add camera: %v", err)
		http.Error(w, "Failed to add camera", http.StatusInternalServerError)
		return
	}

	// newCam.Status = "initializing"
	// s.State.Cameras.Store(newCam.ID, newCam)

	// TODO: Send JSON payload via Stdin to the target C++ Worker Subprocess

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newCamera); err != nil {
		log.Printf("Error encoding new camera response: %v", err)
	}
}

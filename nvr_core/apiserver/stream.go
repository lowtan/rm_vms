package apiserver

import (
	// "context"
	// "encoding/json"
	"log"
	"net/http"
	// "sync"
	// "time"
	"strconv"
	"nvr_core/stream"
)


// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetStream(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))

	if(err != nil) {
		log.Println("[GetStream] failed to get camera id")
		return
	}

	worker := s.PM.CameraWorker(id)
	if(worker == nil) {
		log.Println("[GetStream] failed to get worker")
		return
	}

	hub := worker.StreamHubForCam(id)
	if(hub == nil) {
		log.Println("[GetStream] failed to get stream hub")
		return
	}

	stream.ServeWs(hub, w, r)

}

package apiserver

import (
	"encoding/json"
	"net/http"
)

func (s *APIServer) GetDebugInfo(w http.ResponseWriter, r *http.Request) {

	data, error := s.Services.System.GetDebugData(s.Context)

	if(error != nil) {
		http.Error(w, "failed to get debug info.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
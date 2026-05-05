package apiserver

import (
	"net/http"
)

func (s *APIServer) GetDebugInfo(w http.ResponseWriter, r *http.Request) {

	data, error := s.Services.System.GetDebugData(s.Context)

	if(error != nil) {
		http.Error(w, "failed to get debug info.", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, data)
}
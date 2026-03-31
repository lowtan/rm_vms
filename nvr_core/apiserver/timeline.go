package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"nvr_core/apiserver/dto"
)


// SearchRecords queries the SQLite database for video segments within a time range
func (api *APIServer) GetTimeline(w http.ResponseWriter, r *http.Request) {

	// Parse Query Parameters
	camIDStr := r.PathValue("cam_id")
	startStr := r.PathValue("start")
	endStr := r.PathValue("end")

	camID, err := strconv.Atoi(camIDStr)
	start, err2 := strconv.ParseInt(startStr, 10, 64)
	end, err3 := strconv.ParseInt(endStr, 10, 64)

	if err != nil || err2 != nil || err3 != nil {
		http.Error(w, "Invalid query parameters. Requires cam_id, start, and end (Unix ms)", http.StatusBadRequest)
		return
	}


	svcs := api.Services
	blocks, errTimeline := svcs.Timeline.GetContiguousBlocks(api.Context, camIDStr, start, end)

	if(errTimeline != nil) {
		log.Println("[GetTimeline] failed to get timeline blocks", errTimeline.Error())
		return
	}

	// hostAddr := r.Host // Dynamically grab the server IP/Port

	response := dto.TimelineResponse{
		CameraID: camID,
		Timelines: blocks,
	}

	// Send the JSON payload
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[API] Error encoding search response: %v", err)
	}
}
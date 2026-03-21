package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// --- Struct definitions for the JSON Payload ---

type SegmentItem struct {
	ID         int    `json:"id"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	DurationMs int64  `json:"duration_ms"`
	StreamURL  string `json:"stream_url"`
}

type SearchResponse struct {
	CameraID     int           `json:"camera_id"`
	SearchWindow struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	} `json:"search_window"`
	Segments []SegmentItem `json:"segments"`
}

// --- The Handler ---

// SearchRecords queries the SQLite database for video segments within a time range
func (api *APIServer) SearchRecords(w http.ResponseWriter, r *http.Request) {
	// Parse Query Parameters
	camIDStr := r.URL.Query().Get("cam_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	camID, err := strconv.Atoi(camIDStr)
	start, err2 := strconv.ParseInt(startStr, 10, 64)
	end, err3 := strconv.ParseInt(endStr, 10, 64)

	if err != nil || err2 != nil || err3 != nil {
		http.Error(w, "Invalid query parameters. Requires cam_id, start, and end (Unix ms)", http.StatusBadRequest)
		return
	}

	// TODO: Query your SQLite database here using camID, start, and end.
	// Example mock data representing rows returned from SQLite:
	// rows, _ := db.Query("SELECT id, start_time, end_time FROM segments WHERE...")
	
	hostAddr := r.Host // Dynamically grab the server IP/Port
	
	response := SearchResponse{
		CameraID: camID,
		Segments: []SegmentItem{},
	}
	response.SearchWindow.Start = start
	response.SearchWindow.End = end

	// Map the SQLite rows to the JSON struct
	// (Mocking a single returned row for demonstration)
	mockSegmentID := 42
	mockStart := int64(1710980000000)
	mockEnd := int64(1710980600000)

	segment := SegmentItem{
		ID:         mockSegmentID,
		StartTime:  mockStart,
		EndTime:    mockEnd,
		DurationMs: mockEnd - mockStart,
		// Generate the precise URL for the VLC player
		StreamURL:  fmt.Sprintf("http://%s/api/records/stream?id=%d", hostAddr, mockSegmentID),
	}

	response.Segments = append(response.Segments, segment)

	// Send the JSON payload
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[API] Error encoding search response: %v", err)
	}
}
package dto

// TimelineBlock represents a continuous block of recorded video.
type TimelineBlock struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

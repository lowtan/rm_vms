package dto


type SystemDebugInfo struct {
	Status string `json:"status"`
	TotalSegments int `json:"total_segments"`
	DbSize float64 `json:"db_size_mb"`
	// WalModeActive bool `json:"wal_mode_active"`
	LastSegment SegmentItem `json:"last_segment"`
}

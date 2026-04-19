package shm

type ChannelMetrics struct {
	CamID          int     `json:"cam_id"`
	ChannelID      int     `json:"channel_id"`
	Capacity       uint32  `json:"capacity"`
	Head           uint32  `json:"write_head"`
	Tail           uint32  `json:"read_tail"`
	BytesBuffered  uint32  `json:"bytes_buffered"`
	SaturationPct  float64 `json:"saturation_pct"`
	IsStalled      bool    `json:"is_stalled"`
}

type WorkerMetrics struct {
	WorkerID string                    `json:"worker_id"`
	Channels map[int]*ChannelMetrics   `json:"channels"`
}
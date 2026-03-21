package models

// Camera represents an ONVIF/RTSP video source.
type Camera struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IPAddress       string `json:"ip_address"`
	HTTPPort        int    `json:"http_port"`
	Type            string `json:"type"`
	Username        string `json:"username"`
	PasswordEnc     string `json:"-"` // Never serialize to JSON
	StreamURL       string `json:"stream_url"`
	SubStreamURL    string `json:"sub_stream_url"`
	ONVIFToken      string `json:"onvif_profile_token"`
	SupportsPTZ     bool   `json:"supports_ptz"`
	RetentionGBLimit int    `json:"retention_gb_limit"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

// Segment represents a recorded 1-minute video file.
type Segment struct {
	ID        int64  `json:"id"`
	CameraID  string `json:"camera_id"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	FilePath  string `json:"file_path"`
	SizeBytes int64  `json:"size_bytes"`
}
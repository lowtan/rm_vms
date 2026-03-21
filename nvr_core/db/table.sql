-- cameras, with ONVIF related columns
CREATE TABLE IF NOT EXISTS cameras (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    
    -- Networking & Discovery
    ip_address TEXT NOT NULL,
    http_port INTEGER DEFAULT 80,
    type TEXT NOT NULL DEFAULT 'onvif', -- 'onvif' or 'rtsp_only'

    -- Authentication (Crucial for ONVIF Control Plane)
    username TEXT,
    password_enc TEXT, -- NEVER store plaintext NVR passwords

    -- RTSP Data Plane (Can be dynamically populated via ONVIF)
    stream_url TEXT NOT NULL,
    sub_stream_url TEXT, -- Highly recommended for multi-camera Vue.js grid viewing

    -- ONVIF Specific State
    onvif_profile_token TEXT, -- e.g., 'Profile_1' (Primary high-res stream)
    supports_ptz INTEGER DEFAULT 0,

    -- Storage & State
    retention_gb_limit INTEGER,
    is_active INTEGER DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,   -- Unix timestamp (seconds)
    end_time INTEGER NOT NULL,     -- Unix timestamp (seconds)
    file_path TEXT NOT NULL,       -- e.g., '/storage/cam_01/2026-03-20/19/44_00.mp4'
    size_bytes INTEGER NOT NULL,   -- Tracked for watermark calculations

    FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
);

-- Optimizes: SELECT * FROM segments WHERE camera_id = ? AND start_time >= ? AND end_time <= ?
CREATE INDEX IF NOT EXISTS idx_segments_timeline ON segments(camera_id, start_time, end_time);
-- Optimizes: SELECT * FROM segments ORDER BY start_time ASC LIMIT ?
CREATE INDEX IF NOT EXISTS idx_segments_pruning ON segments(start_time);



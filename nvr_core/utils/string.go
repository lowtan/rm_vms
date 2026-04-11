package utils

import (
	"fmt"
)

func PathForCameraPlayURL(camID string, time int64) string {
	return fmt.Sprintf("/api/cameras/%s/play?time=%d", camID, time)
}

func PathForCameraTSPlayURL(camID string, time int64) string {
	return fmt.Sprintf("/api/cameras/%s/play/ts?time=%d", camID, time)
}

// HandleGetPlaylist expects: GET /api/cameras/{cam_id}/playlist.m3u8?start=1711000000&end=1711003600
func PathForCameraPlaylistURL(camID string, start int64, end int64) string {
	return fmt.Sprintf("/api/cameras/%s/playlist.m3u8?start=%d&end=%d", camID, start, end)
}

func PathForCameraVODPlaylistURL(camID string, start int64, end int64) string {
	return fmt.Sprintf("/api/cameras/%s/playlist/ts.m3u8?start=%d&end=%d", camID, start, end)
}


// For WebSocket live stream
func URLForCameraWSStream(host string, camID string) string {
	return fmt.Sprintf("ws://%s/ws/stream/%s", host, camID)
}
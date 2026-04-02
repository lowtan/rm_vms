package service

import (
	"context"
	"fmt"
	"strings"

	"nvr_core/db/repository"
	"nvr_core/utils"
)

// type M3U8Wrapper struct {
// 	builder strings.Builder
// 	isVOD
// }

type PlaylistService interface {
	// GenerateVODPlaylist creates an M3U8 playlist string for a specific time range.
	// baseURL is injected by the HTTP handler so the service doesn't need to know the server's IP.
	GenerateVODPlaylist(ctx context.Context, camID string, start, end int64, baseURL string) (string, error)
}

func NewPlaylistService(repo repository.SegmentRepository) PlaylistService {
	return &segmentServiceBase{repo: repo}
}

func (s *segmentServiceBase) GenerateVODPlaylist(ctx context.Context, camID string, start, end int64, baseURL string) (string, error) {
	// Fetch all segments within the requested time window
	segments, err := s.repo.GetSegmentsByRange(ctx, camID, start, end)
	if err != nil {
		return "", fmt.Errorf("failed to fetch segments for playlist: %w", err)
	}

	if len(segments) == 0 {
		return "", ErrVideoNotFound // Reusing the error we defined in playback.go
	}

	// Build the M3U8 Header
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n")

	// builder.WriteString("#EXT-X-VERSION:3\n")
	// builder.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	
	// Target duration is the maximum length of any single segment. 
	// Since we cut at 1 minute, we set this safely to 61 seconds to account for keyframe drift.
	// builder.WriteString("#EXT-X-TARGETDURATION:61\n") 
	// builder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n\n")

	// Append each segment
	for _, seg := range segments {
		// Calculate actual duration (e.g., 60.000 seconds)
		durationSeconds := float64(seg.EndTime - seg.StartTime)
		if durationSeconds <= 0 {
			durationSeconds = 60.0 // Failsafe
		}

		// EXTINF defines the length of the upcoming segment
		builder.WriteString(fmt.Sprintf("#EXTINF:%.3f,%d\n", durationSeconds, seg.StartTime))

		// The actual URL VLC will call to get the video bytes.
		// It points directly to our previously built Playback API!
		apiURI := utils.PathForCameraPlayURL(camID, seg.StartTime)
		segmentURL := fmt.Sprintf("%s%s\n", baseURL, apiURI)
		builder.WriteString(segmentURL)
	}

	// Close the playlist
	// builder.WriteString("#EXT-X-ENDLIST\n")

	return builder.String(), nil
}
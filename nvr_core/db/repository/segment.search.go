package repository

import (
	"context"
	"database/sql"
	// "database/sql"
	// "errors"
	// "fmt"
	// "strings"

	"nvr_core/db/models"
)


func (r *segmentRepo) GetSegmentsByRange(ctx context.Context, camID string, start, end int64) ([]*models.Segment, error) {
	query := `
		SELECT id, camera_id, start_time, end_time, file_path, size_bytes 
		FROM segments 
		WHERE camera_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, camID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*models.Segment
	for rows.Next() {
		var seg models.Segment
		if err := rows.Scan(&seg.ID, &seg.CameraID, &seg.StartTime, &seg.EndTime, &seg.FilePath, &seg.SizeBytes); err != nil {
			return nil, err
		}
		segments = append(segments, &seg)
	}
	return segments, rows.Err()
}

func (r *segmentRepo) GetSegmentAtTime(ctx context.Context, camID string, timestamp int64) (*models.Segment, error) {
	// Find the segment where the requested timestamp falls between start and end.
	query := `
		SELECT id, camera_id, start_time, end_time, file_path, size_bytes 
		FROM segments 
		WHERE camera_id = ? AND start_time <= ? AND end_time >= ?
		LIMIT 1
	`

	var seg models.Segment
	err := r.db.QueryRowContext(ctx, query, camID, timestamp, timestamp).Scan(
		&seg.ID, &seg.CameraID, &seg.StartTime, &seg.EndTime, &seg.FilePath, &seg.SizeBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No video exists at this exact second
		}
		return nil, err
	}

	return &seg, nil
}
package repository

import (
	"context"
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


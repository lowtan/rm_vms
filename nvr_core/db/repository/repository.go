package repository

import (
	"context"
	"database/sql"

	"nvr_core/db/models"
)

// SegmentRepository defines the contract for segment data access.
type SegmentRepository interface {
	Insert(ctx context.Context, seg *models.Segment) error
	PruneOldest(ctx context.Context, limit int) ([]string, error)
	IncrementalVacuum(ctx context.Context, pages int) error
}

type segmentRepo struct {
	db *sql.DB
}

func NewSegmentRepository(db *sql.DB) SegmentRepository {
	return &segmentRepo{db: db}
}

func (r *segmentRepo) Insert(ctx context.Context, seg *models.Segment) error {
	query := `INSERT INTO segments (camera_id, start_time, end_time, file_path, size_bytes) 
	          VALUES (?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(ctx, query, seg.CameraID, seg.StartTime, seg.EndTime, seg.FilePath, seg.SizeBytes)
	if err != nil {
		return err
	}
	seg.ID, _ = result.LastInsertId()
	return nil
}

// PruneOldest executes the O(1) eviction policy and returns the physical file paths to delete.
func (r *segmentRepo) PruneOldest(ctx context.Context, limit int) ([]string, error) {
	query := `
		DELETE FROM segments 
		WHERE id IN (
			SELECT id FROM segments ORDER BY start_time ASC LIMIT ?
		)
		RETURNING file_path;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

// IncrementalVacuum reclaims disk space from deleted rows.
func (r *segmentRepo) IncrementalVacuum(ctx context.Context, pages int) error {
	query := `PRAGMA incremental_vacuum(?);`
	_, err := r.db.ExecContext(ctx, query, pages)
	return err
}
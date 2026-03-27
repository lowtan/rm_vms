package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"nvr_core/db/models"
)

/**
 * Sement Basic Operations
 */

// SegmentRepository defines the contract for segment data access.
type SegmentRepository interface {
	Insert(ctx context.Context, seg *models.Segment) error
	PruneOldest(ctx context.Context, limit int) ([]string, error)
	IncrementalVacuum(ctx context.Context, pages int) error
	GetLastSegment(ctx context.Context) (*models.Segment, error)

	// DB/file sanity check
	GetAllFilePaths(ctx context.Context) (map[string]struct{}, error)
	DeleteByFilePaths(ctx context.Context, paths []string) error
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

func (r *segmentRepo) GetLastSegment(ctx context.Context) (*models.Segment, error) {
	query := `
		SELECT id, camera_id, start_time, end_time, file_path, size_bytes 
		FROM segments 
		ORDER BY start_time DESC 
		LIMIT 1
	`
	
	var seg models.Segment
	err := r.db.QueryRowContext(ctx, query).Scan(
		&seg.ID,
		&seg.CameraID,
		&seg.StartTime,
		&seg.EndTime,
		&seg.FilePath,
		&seg.SizeBytes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil gracefully if the database is completely empty
		}
		return nil, err
	}

	return &seg, nil
}

/**
 * DB/file sanity check
 */

// GetAllFilePaths returns a Hash Set of all file paths currently tracked in the DB.
func (r *segmentRepo) GetAllFilePaths(ctx context.Context) (map[string]struct{}, error) {
	query := `SELECT file_path FROM segments`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query file paths: %w", err)
	}
	defer rows.Close()

	// Using an empty struct{} consumes 0 bytes of memory in Go
	paths := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths[path] = struct{}{}
	}
	return paths, rows.Err()
}

// DeleteByFilePaths removes ghost records in batches.
func (r *segmentRepo) DeleteByFilePaths(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// Batch delete in chunks of 900 to respect SQLite's variable limit
	chunkSize := 900
	for i := 0; i < len(paths); i += chunkSize {
		end := i + chunkSize
		if end > len(paths) {
			end = len(paths)
		}
		chunk := paths[i:end]

		placeholders := strings.Repeat("?,", len(chunk)-1) + "?"
		query := fmt.Sprintf(`DELETE FROM segments WHERE file_path IN (%s)`, placeholders)

		args := make([]interface{}, len(chunk))
		for j, p := range chunk {
			args[j] = p
		}

		if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to delete ghost records: %w", err)
		}
	}
	return nil
}
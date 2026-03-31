package service

import (
	"context"
	"database/sql"

	"nvr_core/apiserver/dto"
	"nvr_core/db/repository"
)

type SystemService interface {
	GetDebugInfo(ctx context.Context) (dto.SystemDebugInfo, error)
}


func NewSystemService(db *sql.DB, repo repository.SegmentRepository) SystemService {
	return &debugServiceBase{db: db, repo: repo}
}

func (s *debugServiceBase) GetDebugInfo(ctx context.Context) (dto.SystemDebugInfo, error) {

	// Query SQLite internal stats
	var totalSegments int
	var dbSizeBytes int64

	// Count total rows
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM segments").Scan(&totalSegments)

	// Check the physical file size of the database itself using SQLite pragmas
	s.db.QueryRowContext(ctx, "SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()").Scan(&dbSizeBytes)

	segment, err := s.repo.GetLastSegment(ctx)
	if err != nil {
		return dto.SystemDebugInfo{}, err
	}

	if segment == nil {
		return dto.SystemDebugInfo{}, nil
	}

	// Start the first block
	data := dto.SystemDebugInfo{
		Status: "online",
		TotalSegments: totalSegments,
		DbSize: float64(dbSizeBytes) / 1024 / 1024,
		LastSegment: dto.NewSegmentItemFrom(segment),
	}

	return data, nil
}
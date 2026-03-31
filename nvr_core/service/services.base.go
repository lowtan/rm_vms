package service

import (
	"database/sql"
	"nvr_core/db/repository"
)

type segmentServiceBase struct {
	repo repository.SegmentRepository
}

type debugServiceBase struct {
	db *sql.DB
	repo repository.SegmentRepository
}

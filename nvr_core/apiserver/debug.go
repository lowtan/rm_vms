package apiserver

import (
	// "context"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"nvr_core/db/repository"
)

type DebugHandler struct {
	ctx context.Context
	db *sql.DB
	segment repository.SegmentRepository
}

func NewDebugHandler(ctx context.Context, db *sql.DB, segment repository.SegmentRepository) *DebugHandler {
	return &DebugHandler{db: db, ctx: ctx, segment: segment}
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Query SQLite internal stats
	var totalSegments int
	var dbSizeBytes int64

	// Count total rows
	h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM segments").Scan(&totalSegments)

	// Check the physical file size of the database itself using SQLite pragmas
	h.db.QueryRowContext(r.Context(), "SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()").Scan(&dbSizeBytes)

	// Fetch the last segment via the Repository
	lastSegment, err := h.segment.GetLastSegment(h.ctx)
	if err != nil {
		log.Printf("[API] Failed to fetch last segment: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"status":          "online",
		"total_segments":  totalSegments,
		"db_size_mb":      float64(dbSizeBytes) / 1024 / 1024,
		"wal_mode_active": true,
		"last_segment":    lastSegment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
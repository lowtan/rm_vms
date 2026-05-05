package dto

type DailySummary struct {
	Date         string `json:"date"`          // e.g., "2026-05-01"
	TotalSeconds int    `json:"total_seconds"` // e.g., 86400
	// Formatted    string `json:"formatted"`     // e.g., "24:00:00"
}
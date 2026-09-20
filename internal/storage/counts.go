package storage

import (
	"time"

	"gorm.io/gorm"
)

// GameCounts contains only aggregate service activity, never game identifiers.
type GameCounts struct {
	Active    int64 `json:"active"`
	Completed int64 `json:"completed"`
}

func CountGames(db *gorm.DB, cutoff time.Time) (GameCounts, error) {
	var counts GameCounts
	err := db.Model(&Game{}).Select(`
  COUNT(*) FILTER (WHERE ended_at IS NULL AND COALESCE(result, '') NOT IN ('1-0', '0-1', '1/2-1/2') AND updated_at >= ?) AS active,
  COUNT(*) FILTER (WHERE ended_at IS NOT NULL OR result IN ('1-0', '0-1', '1/2-1/2')) AS completed`, cutoff).Scan(&counts).Error
	return counts, err
}

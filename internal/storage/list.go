package storage

import (
	"time"

	"gorm.io/gorm"
)

// PublicGame deliberately excludes seat credentials and private recovery data.
type PublicGame struct {
	ID        string    `json:"id"`
	Result    string    `json:"result"`
	MoveCount int       `json:"moveCount"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ListGames(db *gorm.DB, completed bool, cutoff time.Time, offset, limit int) ([]PublicGame, bool, error) {
	rows := make([]PublicGame, 0)
	query := db.Model(&Game{}).Select("id, result, move_count, updated_at")
	if completed {
		query = query.Where("ended_at IS NOT NULL OR result IN ?", []string{"1-0", "0-1", "1/2-1/2"})
	} else {
		query = query.Where("ended_at IS NULL AND COALESCE(result, '') NOT IN ? AND updated_at >= ?", []string{"1-0", "0-1", "1/2-1/2"}, cutoff)
	}
	if err := query.Order("updated_at DESC, id ASC").Offset(offset).Limit(limit + 1).Scan(&rows).Error; err != nil {
		return nil, false, err
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	}
	return rows, more, nil
}

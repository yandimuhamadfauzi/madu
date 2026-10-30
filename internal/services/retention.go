package services

import (
	"time"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

// PurgeOldEvents menghapus log serangan yang lebih lama dari batas retensi (IP bisa dianggap data pribadi).
func PurgeOldEvents(db *gorm.DB, days int) {
	db.Where("created_at < ?", time.Now().AddDate(0, 0, -days)).Delete(&models.Event{})
}

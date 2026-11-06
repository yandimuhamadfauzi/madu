package services

import (
	"time"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

// AutoBlockIP memasukkan IP ke blocklist untuk durasi d. Hasilnya true bila blokir baru dibuat atau
// diperbarui. Blokir manual (permanen) dan blokir otomatis yang masih aktif tidak diubah.
func AutoBlockIP(db *gorm.DB, ip, reason string, d time.Duration) (bool, error) {
	now := time.Now()
	exp := now.Add(d)

	var rows []models.Blocklist
	if err := db.Where("ip = ?", ip).Limit(1).Find(&rows).Error; err != nil {
		return false, err
	}
	if len(rows) == 0 {
		if err := db.Create(&models.Blocklist{IP: ip, Reason: reason, Auto: true, ExpiresAt: &exp}).Error; err != nil {
			return false, err
		}
		Invalidate()
		return true, nil
	}

	b := rows[0]
	if b.ExpiresAt == nil || b.ExpiresAt.After(now) {
		return false, nil // permanen/manual, atau masih aktif
	}
	// blokir otomatis lama yang sudah kedaluwarsa: perbarui
	err := db.Model(&models.Blocklist{}).Where("id = ?", b.ID).Updates(map[string]any{
		"reason": reason, "auto": true, "expires_at": exp, "created_at": now,
	}).Error
	if err != nil {
		return false, err
	}
	Invalidate()
	return true, nil
}

// PurgeExpiredBlocks menghapus blokir otomatis yang sudah kedaluwarsa.
func PurgeExpiredBlocks(db *gorm.DB) {
	db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&models.Blocklist{})
}

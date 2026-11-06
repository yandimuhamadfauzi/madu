package services

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

const settingAutoBlock = "autoblock_enabled"

// ErrAutoBlockLocked berarti blokir otomatis dimatikan lewat environment (AUTOBLOCK_ENABLED=false).
var ErrAutoBlockLocked = errors.New("blokir otomatis dikunci lewat environment")

// LoadAutoBlockSetting menerapkan pilihan admin yang tersimpan di database. Bila belum pernah diatur,
// nilai bawaan dari environment tetap dipakai.
func LoadAutoBlockSetting(db *gorm.DB, a *AutoBlocker) {
	var rows []models.Setting
	if err := db.Where("name = ?", settingAutoBlock).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return
	}
	a.SetEnabled(rows[0].Value == "1")
}

// SaveAutoBlockSetting menyimpan pilihan admin ke database lalu menerapkannya langsung.
func SaveAutoBlockSetting(db *gorm.DB, a *AutoBlocker, on bool) error {
	if a.Locked() {
		return ErrAutoBlockLocked
	}
	v := "0"
	if on {
		v = "1"
	}
	err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&models.Setting{Name: settingAutoBlock, Value: v}).Error
	if err != nil {
		return err
	}
	a.SetEnabled(on)
	return nil
}

// FlushAutoBlocks melepas semua blokir otomatis (blokir manual tidak disentuh) dan mengembalikan jumlahnya.
func FlushAutoBlocks(db *gorm.DB) int64 {
	res := db.Where("auto = ?", true).Delete(&models.Blocklist{})
	Invalidate()
	return res.RowsAffected
}

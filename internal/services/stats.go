package services

import (
	"time"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Count struct {
	Label string
	N     int64
}

type Stats struct {
	Total, UniqueIPs, High, Blocked  int64
	ByType, ByHour, TopPaths, TopIPs []Count
	MaxType, MaxHour, MaxPath, MaxIP int64
}

func events(db *gorm.DB, withTest bool) *gorm.DB {
	q := db.Model(&models.Event{})
	if !withTest {
		q = q.Where("is_test = ?", false)
	}
	return q
}

func maxOf(c []Count) int64 {
	var m int64 = 1
	for _, x := range c {
		if x.N > m {
			m = x.N
		}
	}
	return m
}

func GetStats(db *gorm.DB, withTest bool) Stats {
	var s Stats
	events(db, withTest).Count(&s.Total)
	events(db, withTest).Distinct("ip").Count(&s.UniqueIPs)
	events(db, withTest).Where("severity = ?", 3).Count(&s.High)
	events(db, withTest).Where("blocked = ?", true).Count(&s.Blocked)

	events(db, withTest).Select("attack_type AS label, COUNT(*) AS n").
		Group("attack_type").Order("n DESC").Scan(&s.ByType)

	since := time.Now().Add(-24 * time.Hour)
	events(db, withTest).Select("DATE_FORMAT(created_at, '%H:00') AS label, COUNT(*) AS n").
		Where("created_at >= ?", since).Group("label").Order("MIN(created_at)").Scan(&s.ByHour)

	// CHAR(63) adalah tanda tanya; dipakai agar tidak dianggap placeholder oleh GORM.
	events(db, withTest).Select("SUBSTRING_INDEX(path, CHAR(63), 1) AS label, COUNT(*) AS n").
		Group("label").Order("n DESC").Limit(5).Scan(&s.TopPaths)

	events(db, withTest).Select("ip AS label, COUNT(*) AS n").
		Group("ip").Order("n DESC").Limit(5).Scan(&s.TopIPs)

	s.MaxType, s.MaxHour, s.MaxPath, s.MaxIP = maxOf(s.ByType), maxOf(s.ByHour), maxOf(s.TopPaths), maxOf(s.TopIPs)
	return s
}

package services

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

const cacheTTL = 20 * time.Second

type snapshot struct {
	rules   []models.Rule
	blocked map[string]struct{}
	at      time.Time
}

var (
	cacheMu sync.RWMutex
	cache   *snapshot
)

// Invalidate dipanggil setelah admin mengubah aturan atau blocklist.
func Invalidate() {
	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()
}

func current(db *gorm.DB) *snapshot {
	cacheMu.RLock()
	c := cache
	cacheMu.RUnlock()
	if c != nil && time.Since(c.at) < cacheTTL {
		return c
	}
	s := &snapshot{blocked: map[string]struct{}{}, at: time.Now()}
	db.Where("active = ?", true).Order("id").Find(&s.rules)
	var ips []string
	db.Model(&models.Blocklist{}).Pluck("ip", &ips)
	for _, ip := range ips {
		s.blocked[ip] = struct{}{}
	}
	cacheMu.Lock()
	cache = s
	cacheMu.Unlock()
	return s
}

func ActiveRules(db *gorm.DB) []models.Rule { return current(db).rules }

func IsBlocked(db *gorm.DB, ip string) bool {
	_, ok := current(db).blocked[ip]
	return ok
}

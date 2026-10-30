package handlers

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Export struct{ DB *gorm.DB }

// csvSafe menetralkan formula injection: sel yang diawali = + - @ dianggap rumus oleh Excel.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func yesNo(b bool) string {
	if b {
		return "ya"
	}
	return "tidak"
}

func (e *Export) CSV(c *gin.Context) {
	tx := e.DB.Model(&models.Event{})
	if c.Query("uji") != "1" {
		tx = tx.Where("is_test = ?", false)
	}
	var evs []models.Event
	tx.Order("id DESC").Limit(5000).Find(&evs)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="madu-serangan.csv"`)
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"waktu", "ip", "method", "path", "user_agent", "jenis", "tingkat", "diblokir", "uji"})
	for _, ev := range evs {
		_ = w.Write([]string{
			ev.CreatedAt.Format("2006-01-02 15:04:05"), csvSafe(ev.IP), csvSafe(ev.Method), csvSafe(ev.Path),
			csvSafe(ev.UserAgent), csvSafe(ev.AttackType), strconv.Itoa(ev.Severity), yesNo(ev.Blocked), yesNo(ev.IsTest),
		})
	}
	w.Flush()
}

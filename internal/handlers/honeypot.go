package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

type Honeypot struct {
	DB      *gorm.DB
	Limiter *services.FailLimiter // batas pencatatan per IP
	Global  *services.FailLimiter // batas pencatatan seluruh trafik (melindungi database)
}

// clip membuang byte UTF-8 tidak valid (MySQL menolaknya) dan membatasi panjang data dari penyerang.
func clip(s string, n int) string {
	s = strings.ToValidUTF8(s, "")
	if len(s) > n {
		s = strings.ToValidUTF8(s[:n], "")
	}
	return s
}

func (h *Honeypot) Handle(c *gin.Context) {
	if c.Request.URL.Path == "/favicon.ico" {
		c.Status(http.StatusNotFound)
		return
	}
	ip := c.ClientIP()
	raw := c.Request.URL.RequestURI() // path + query. Body request TIDAK dibaca dan TIDAK disimpan
	ua := c.Request.UserAgent()

	blocked := services.IsBlocked(h.DB, ip)
	rules := services.ActiveRules(h.DB)
	rule := services.Classify(raw, rules)
	if rule == nil {
		rule = services.Classify(ua, rules) // alat pemindai sering terbaca dari user-agent
	}

	// Pencatatan dibatasi per IP dan secara global agar database tidak bisa dibanjiri.
	if h.Limiter.Blocked(ip) == 0 && h.Global.Blocked("semua") == 0 {
		h.Limiter.Fail(ip)
		h.Global.Fail("semua")
		ev := models.Event{
			IP:         ip,
			Method:     clip(c.Request.Method, 10),
			Path:       clip(raw, 500),
			UserAgent:  clip(ua, 300),
			AttackType: "Tidak dikenal",
			Severity:   1,
			Blocked:    blocked,
			IsTest:     ip == "127.0.0.1" || ip == "::1",
		}
		if rule != nil {
			ev.AttackType = rule.AttackType
			ev.Severity = rule.Severity
		}
		h.DB.Create(&ev)
	}

	switch {
	case blocked:
		c.String(http.StatusForbidden, "Forbidden")
	case rule != nil:
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body></body></html>")) // umpan
	default:
		c.String(http.StatusNotFound, "404 page not found")
	}
}

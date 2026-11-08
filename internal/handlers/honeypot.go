package handlers

import (
	"fmt"
	"html/template" // Tambahan untuk memproses HTML
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

type Honeypot struct {
	DB        *gorm.DB
	Limiter   *services.FailLimiter // batas pencatatan per IP
	Global    *services.FailLimiter // batas pencatatan seluruh trafik (melindungi database)
	AutoBlock *services.AutoBlocker
}

// clip membuang byte UTF-8 tidak valid (MySQL menolaknya) dan membatasi panjang data dari penyerang.
func clip(s string, n int) string {
	s = strings.ToValidUTF8(s, "")
	if len(s) > n {
		s = strings.ToValidUTF8(s[:n], "")
	}
	return s
}

// autoBlock menghitung serangan berat per IP dan memblokir IP saat ambang tercapai.
func (h *Honeypot) autoBlock(ip string, rule *models.Rule) {
	if rule == nil || !h.AutoBlock.Record(ip, rule.Severity) {
		return
	}
	cfg := h.AutoBlock.Config()
	reason := fmt.Sprintf("Otomatis: %d serangan berat dalam %d menit", cfg.Threshold, int(cfg.Window/time.Minute))
	ok, err := services.AutoBlockIP(h.DB, ip, reason, cfg.Duration)
	if err != nil {
		log.Println("autoblock gagal:", err)
		return
	}
	if ok {
		h.DB.Create(&models.AuditLog{
			Actor: "system", Action: "autoblock", IP: ip,
			Detail: fmt.Sprintf("IP diblokir otomatis selama %d menit", int(cfg.Duration/time.Minute)),
		})
	}
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
	if !blocked {
		h.autoBlock(ip, rule)
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

	// LOGIKA PENANGANAN RESPON BERDASARKAN STATUS
	switch {
	case blocked:
		c.String(http.StatusForbidden, "Forbidden")
		
	// --- TAMBAHAN JEBAKAN LOGIN WORDPRESS ---
	case c.Request.URL.Path == "/wp-login.php" || c.Request.URL.Path == "/wp-admin":
		tmplPath := "web/templates/wp_login.html"
		
		// Jika bot mengirimkan form kredensial (POST)
		if c.Request.Method == http.MethodPost {
			username := clip(c.PostForm("log"), 100)
			password := clip(c.PostForm("pwd"), 100)
			
			// Tampilkan di terminal / log (bisa dimodifikasi untuk simpan ke DB nanti)
			log.Printf("🚨 [TRAP WP-LOGIN] IP: %s | Credential Captured -> User: '%s' | Pass: '%s'\n", ip, username, password)

			tmpl, err := template.ParseFiles(tmplPath)
			if err != nil {
				c.String(http.StatusInternalServerError, "Server Error")
				return
			}
			
			c.Status(http.StatusForbidden) // Status ditolak agar bot tahu gagal
			tmpl.Execute(c.Writer, map[string]interface{}{
				"Error":    true,
				"Username": username,
			})
			return
		}

		// Jika bot hanya mengakses halaman biasa (GET)
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			c.String(http.StatusInternalServerError, "Server Error")
			return
		}
		
		c.Status(http.StatusOK)
		tmpl.Execute(c.Writer, nil)
	// ----------------------------------------

	case rule != nil:
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body></body></html>")) // umpan
	default:
		c.String(http.StatusNotFound, "404 page not found")
	}
}
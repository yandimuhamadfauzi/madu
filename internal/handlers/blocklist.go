package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

type Blocks struct {
	DB   *gorm.DB
	Auto *services.AutoBlocker
}

func (b *Blocks) page(c *gin.Context, status int, msg string) {
	u := c.MustGet("user").(*models.User)
	var items []models.Blocklist
	b.DB.Where("expires_at IS NULL OR expires_at > ?", time.Now()).Order("id DESC").Find(&items)
	c.HTML(status, "blocklist.html", gin.H{
		"Title": "Blocklist", "Username": u.Username, "Role": u.Role, "CSRF": c.GetString("csrf"),
		"Items": items, "Error": msg,
		"AutoOn": b.Auto.Enabled(), "AutoLocked": b.Auto.Locked(),
		"AutoThreshold": b.Auto.Config().Threshold, "AutoSeverity": b.Auto.Config().MinSeverity,
		"AutoWindowMin": int(b.Auto.Config().Window / time.Minute), "AutoDurationMin": int(b.Auto.Config().Duration / time.Minute),
	})
}

func (b *Blocks) List(c *gin.Context) { b.page(c, http.StatusOK, "") }

func (b *Blocks) Add(c *gin.Context) {
	ip := net.ParseIP(strings.TrimSpace(c.PostForm("ip")))
	reason := strings.TrimSpace(c.PostForm("alasan"))
	if ip == nil {
		b.page(c, http.StatusBadRequest, "Alamat IP tidak valid.")
		return
	}
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) > 100 {
		b.page(c, http.StatusBadRequest, "Alasan maksimal 100 karakter.")
		return
	}

	var rows []models.Blocklist
	if err := b.DB.Where("ip = ?", ip.String()).Limit(1).Find(&rows).Error; err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	if len(rows) > 0 {
		cur := rows[0]
		active := cur.ExpiresAt == nil || cur.ExpiresAt.After(time.Now())
		if active && !cur.Auto {
			b.page(c, http.StatusBadRequest, "IP sudah ada di daftar blokir.")
			return
		}
		// Blokir otomatis (aktif atau sudah kedaluwarsa) dijadikan permanen oleh admin.
		err := b.DB.Model(&models.Blocklist{}).Where("id = ?", cur.ID).Updates(map[string]any{
			"reason": reason, "auto": false, "expires_at": nil, "created_at": time.Now(),
		}).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "terjadi kesalahan")
			return
		}
		c.Redirect(http.StatusSeeOther, "/panel/blokir")
		return
	}
	if err := b.DB.Create(&models.Blocklist{IP: ip.String(), Reason: reason}).Error; err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	c.Redirect(http.StatusSeeOther, "/panel/blokir")
}

func (b *Blocks) Delete(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	b.DB.Where("id = ?", id).Delete(&models.Blocklist{})
	c.Redirect(http.StatusSeeOther, "/panel/blokir")
}

// ExportTXT menghasilkan daftar IP terblokir dalam format teks biasa untuk Firewall (Mikrotik/Cloudflare)
func (b *Blocks) ExportTXT(c *gin.Context) {
	var items []models.Blocklist
	
	// Ambil semua IP yang masih diblokir (baik permanen maupun otomatis yang belum expired)
	if err := b.DB.Where("expires_at IS NULL OR expires_at > ?", time.Now()).Find(&items).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengambil data blocklist")
		return
	}

	// Set header agar dibaca sebagai teks biasa
	c.Header("Content-Type", "text/plain; charset=utf-8")
	
	// Tulis header informasi
	c.String(http.StatusOK, "# Daftar DPO (Blocklist) Madu Honeypot\n")
	c.String(http.StatusOK, "# Digenerate pada: %s\n", time.Now().Format(time.RFC3339))
	c.String(http.StatusOK, "# Total IP: %d\n\n", len(items))

	// Cetak setiap IP di baris baru
	for _, item := range items {
		c.String(http.StatusOK, "%s\n", item.IP)
	}
}

// CheckIPAbuse merespons permintaan informasi reputasi IP dari frontend
func (b *Blocks) CheckIPAbuse(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter IP diperlukan"})
		return
	}

	// Panggil service AbuseIPDB
	info, err := services.CheckIPReputation(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kembalikan data intelijen ke frontend dalam format JSON
	c.JSON(http.StatusOK, info.Data)
}
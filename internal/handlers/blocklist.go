package handlers

import (
	"net"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Blocks struct{ DB *gorm.DB }

func (b *Blocks) page(c *gin.Context, status int, msg string) {
	u := c.MustGet("user").(*models.User)
	var items []models.Blocklist
	b.DB.Order("id DESC").Find(&items)
	c.HTML(status, "blocklist.html", gin.H{
		"Title": "Blocklist", "Username": u.Username, "Role": u.Role, "CSRF": c.GetString("csrf"),
		"Items": items, "Error": msg,
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
	var n int64
	b.DB.Model(&models.Blocklist{}).Where("ip = ?", ip.String()).Count(&n)
	if n > 0 {
		b.page(c, http.StatusBadRequest, "IP sudah ada di daftar blokir.")
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

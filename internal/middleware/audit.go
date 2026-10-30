package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

var actorRe = regexp.MustCompile(`^[A-Za-z0-9]{3,20}$`)

func clipStr(s string, n int) string {
	s = strings.ToValidUTF8(s, "")
	if len(s) > n {
		s = strings.ToValidUTF8(s[:n], "")
	}
	return s
}

func write(db *gorm.DB, c *gin.Context, actor, action string) {
	db.Create(&models.AuditLog{
		Actor:  actor,
		Action: action,
		Detail: clipStr(c.Request.Method+" "+c.Request.URL.Path, 200), // path saja, isi form tidak dicatat
		IP:     c.ClientIP(),
		Status: c.Writer.Status(),
	})
}

// Audit mencatat aksi pengguna yang sudah login, termasuk yang ditolak (403).
func Audit(db *gorm.DB, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		actor := "-"
		if v, ok := c.Get("user"); ok {
			actor = v.(*models.User).Username
		}
		write(db, c, actor, action)
	}
}

// AuditLogin mencatat percobaan login. Username hanya dicatat bila formatnya valid.
func AuditLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.PostForm("username")
		c.Next()
		status := c.Writer.Status()
		if status == http.StatusTooManyRequests {
			return // sudah dibatasi rate limiter, tidak dicatat agar tabel tidak bisa dibanjiri
		}
		if !actorRe.MatchString(u) {
			u = "?"
		}
		action := "login.gagal"
		if status == http.StatusSeeOther {
			action = "login.berhasil"
		}
		write(db, c, u, action)
	}
}

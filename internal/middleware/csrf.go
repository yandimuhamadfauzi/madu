package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireCSRF dipasang setelah RequireAuth pada semua rute POST yang mengubah data.
func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		want := c.GetString("csrf")
		got := c.PostForm("csrf")
		if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
			c.String(http.StatusForbidden, "permintaan ditolak (token CSRF tidak valid)")
			c.Abort()
			return
		}
		c.Next()
	}
}

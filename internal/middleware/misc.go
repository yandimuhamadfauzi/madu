package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandimuhamadfauzi/madu/internal/services"
)

// LimitBody membatasi ukuran body yang boleh dibaca dari satu permintaan.
func LimitBody(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}

// CacheBust membuang cache aturan dan blocklist setelah permintaan perubahan selesai.
func CacheBust() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		services.Invalidate()
	}
}

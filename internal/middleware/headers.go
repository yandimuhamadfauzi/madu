package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders(hsts bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		
		// CSP SUPER AMAN: Mengizinkan domain Esri untuk memuat gambar peta
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com; img-src 'self' data: https://*.tile.openstreetmap.org https://server.arcgisonline.com; connect-src 'self' https://get.geojs.io; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if hsts {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if strings.HasPrefix(c.Request.URL.Path, "/panel") || c.Request.URL.Path == "/masuk" || c.Request.URL.Path == "/daftar" {
			h.Set("Cache-Control", "no-store") 
		}
		c.Next()
	}
}
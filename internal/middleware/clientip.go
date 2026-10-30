package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizeClientIP memastikan header IP dari platform hosting hanya berisi alamat IP yang valid.
func SanitizeClientIP(header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if header != "" {
			ip := net.ParseIP(strings.TrimSpace(c.GetHeader(header)))
			if ip == nil {
				c.Request.Header.Del(header)
			} else {
				c.Request.Header.Set(header, ip.String())
			}
		}
		c.Next()
	}
}

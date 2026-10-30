package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

// RequireRole dipasang setelah RequireAuth. Pengecekan dilakukan di server.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.MustGet("user").(*models.User)
		for _, r := range roles {
			if u.Role == r {
				c.Next()
				return
			}
		}
		c.String(http.StatusForbidden, "akses ditolak")
		c.Abort()
	}
}

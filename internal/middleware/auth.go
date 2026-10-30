package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/services"
)

func RequireAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, err := c.Cookie(services.SessionCookie)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/masuk")
			c.Abort()
			return
		}
		u, s, err := services.SessionFromToken(db, t)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/masuk")
			c.Abort()
			return
		}
		c.Set("user", u)
		c.Set("csrf", s.CSRFToken)
		c.Next()
	}
}

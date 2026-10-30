package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Audit struct{ DB *gorm.DB }

func (a *Audit) List(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	var items []models.AuditLog
	a.DB.Order("id DESC").Limit(100).Find(&items)
	c.HTML(http.StatusOK, "audit.html", gin.H{
		"Title": "Audit log", "Username": u.Username, "Role": u.Role, "CSRF": c.GetString("csrf"), "Items": items,
	})
}

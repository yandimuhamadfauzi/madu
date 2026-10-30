package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminArea(c *gin.Context) {
	c.String(http.StatusOK, "Area admin: hanya admin yang bisa melihat halaman ini.")
}

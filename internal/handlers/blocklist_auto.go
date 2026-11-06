package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandimuhamadfauzi/madu/internal/services"
)

func (b *Blocks) setAuto(c *gin.Context, on bool) {
	if err := services.SaveAutoBlockSetting(b.DB, b.Auto, on); err != nil {
		if errors.Is(err, services.ErrAutoBlockLocked) {
			b.page(c, http.StatusBadRequest, "Blokir otomatis dimatikan lewat environment (AUTOBLOCK_ENABLED=false) dan tidak bisa diubah dari web.")
			return
		}
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	c.Redirect(http.StatusSeeOther, "/panel/blokir")
}

// AutoOn menyalakan blokir otomatis.
func (b *Blocks) AutoOn(c *gin.Context) { b.setAuto(c, true) }

// AutoOff mematikan blokir otomatis. Blokir yang sudah ada tetap berlaku sampai berakhir atau dilepas.
func (b *Blocks) AutoOff(c *gin.Context) { b.setAuto(c, false) }

// AutoFlush melepas semua blokir otomatis sekaligus (blokir manual tidak disentuh).
func (b *Blocks) AutoFlush(c *gin.Context) {
	services.FlushAutoBlocks(b.DB)
	c.Redirect(http.StatusSeeOther, "/panel/blokir")
}

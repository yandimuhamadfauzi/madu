package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Users struct{ DB *gorm.DB }

func (h *Users) page(c *gin.Context, status int, msg string) {
	me := c.MustGet("user").(*models.User)
	var list []models.User
	h.DB.Order("CASE status WHEN 'menunggu' THEN 0 ELSE 1 END, id").Find(&list)
	c.HTML(status, "users.html", gin.H{
		"Title": "Pengguna", "Username": me.Username, "Role": me.Role, "CSRF": c.GetString("csrf"),
		"Users": list, "MeID": me.ID, "Error": msg,
	})
}

func (h *Users) List(c *gin.Context) { h.page(c, http.StatusOK, "") }

// target memuat pengguna dari :id. Aksi terhadap akun sendiri ditolak,
// sehingga admin tidak bisa mengunci dirinya sendiri dan selalu ada minimal satu admin aktif.
func (h *Users) target(c *gin.Context) (*models.User, bool) {
	id, ok := paramID(c)
	if !ok {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return nil, false
	}
	var u models.User
	if err := h.DB.First(&u, id).Error; err != nil {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return nil, false
	}
	me := c.MustGet("user").(*models.User)
	if me.ID == u.ID {
		h.page(c, http.StatusBadRequest, "Tidak bisa mengubah akun sendiri.")
		return nil, false
	}
	return &u, true
}

func (h *Users) setStatus(c *gin.Context, from []string, to string) {
	u, ok := h.target(c)
	if !ok {
		return
	}
	allowed := false
	for _, f := range from {
		if u.Status == f {
			allowed = true
		}
	}
	if !allowed {
		h.page(c, http.StatusBadRequest, "Status akun tidak sesuai untuk aksi ini.")
		return
	}
	h.DB.Model(u).Update("status", to)
	if to != "aktif" {
		h.DB.Where("user_id = ?", u.ID).Delete(&models.Session{}) // cabut semua sesi aktifnya
	}
	c.Redirect(http.StatusSeeOther, "/panel/pengguna")
}

func (h *Users) Approve(c *gin.Context)    { h.setStatus(c, []string{"menunggu"}, "aktif") }
func (h *Users) Deactivate(c *gin.Context) { h.setStatus(c, []string{"aktif"}, "nonaktif") }
func (h *Users) Activate(c *gin.Context)   { h.setStatus(c, []string{"nonaktif"}, "aktif") }

func (h *Users) SetRole(c *gin.Context) {
	u, ok := h.target(c)
	if !ok {
		return
	}
	role := c.PostForm("peran")
	if role != "admin" && role != "analis" {
		h.page(c, http.StatusBadRequest, "Peran tidak valid.")
		return
	}
	h.DB.Model(u).Update("role", role)
	c.Redirect(http.StatusSeeOther, "/panel/pengguna")
}

func (h *Users) Delete(c *gin.Context) {
	u, ok := h.target(c)
	if !ok {
		return
	}
	if u.Status == "aktif" {
		h.page(c, http.StatusBadRequest, "Nonaktifkan akun terlebih dahulu sebelum menghapus.")
		return
	}
	h.DB.Where("user_id = ?", u.ID).Delete(&models.Session{})
	h.DB.Delete(&models.User{}, u.ID)
	c.Redirect(http.StatusSeeOther, "/panel/pengguna")
}

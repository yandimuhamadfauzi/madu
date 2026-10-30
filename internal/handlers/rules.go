package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

type Rules struct{ DB *gorm.DB }

// Jenis yang boleh dipilih untuk aturan (tanpa "Tidak dikenal", yang paling akhir di daftar).
func ruleTypes() []string { return AttackTypes[:len(AttackTypes)-1] }

func paramID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func (r *Rules) page(c *gin.Context, status int, msg string) {
	u := c.MustGet("user").(*models.User)
	var rules []models.Rule
	r.DB.Order("id").Find(&rules)
	var edit *models.Rule
	if id, err := strconv.ParseUint(c.Query("ubah"), 10, 32); err == nil && u.Role == "admin" {
		for i := range rules {
			if uint64(rules[i].ID) == id {
				edit = &rules[i]
			}
		}
	}
	c.HTML(status, "rules.html", gin.H{
		"Title": "Aturan jebakan", "Username": u.Username, "Role": u.Role, "CSRF": c.GetString("csrf"),
		"Rules": rules, "Types": ruleTypes(), "Edit": edit, "Error": msg,
	})
}

func (r *Rules) List(c *gin.Context) { r.page(c, http.StatusOK, "") }

// parseRule memvalidasi semua input dari form di sisi server.
func parseRule(c *gin.Context) (models.Rule, string) {
	name := strings.TrimSpace(c.PostForm("nama"))
	pattern := strings.ToLower(strings.TrimSpace(c.PostForm("pola")))
	typ := c.PostForm("jenis")
	sev, _ := strconv.Atoi(c.PostForm("tingkat"))

	if !utf8.ValidString(name) || !utf8.ValidString(pattern) {
		return models.Rule{}, "Karakter tidak valid."
	}
	if n := utf8.RuneCountInString(name); n < 3 || n > 40 {
		return models.Rule{}, "Nama harus 3-40 karakter."
	}
	if n := utf8.RuneCountInString(pattern); n < 2 || n > 40 {
		return models.Rule{}, "Pola harus 2-40 karakter."
	}
	okType := false
	for _, t := range ruleTypes() {
		if t == typ {
			okType = true
		}
	}
	if !okType {
		return models.Rule{}, "Jenis tidak valid."
	}
	if sev < 1 || sev > 3 {
		return models.Rule{}, "Tingkat tidak valid."
	}
	return models.Rule{Name: name, Pattern: pattern, AttackType: typ, Severity: sev}, ""
}

func (r *Rules) Create(c *gin.Context) {
	rule, msg := parseRule(c)
	if msg != "" {
		r.page(c, http.StatusBadRequest, msg)
		return
	}
	rule.Active = true
	if err := r.DB.Create(&rule).Error; err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	c.Redirect(http.StatusSeeOther, "/panel/aturan")
}

func (r *Rules) Update(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	var cur models.Rule
	if err := r.DB.First(&cur, id).Error; err != nil {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	rule, msg := parseRule(c)
	if msg != "" {
		r.page(c, http.StatusBadRequest, msg)
		return
	}
	r.DB.Model(&cur).Updates(map[string]any{
		"name": rule.Name, "pattern": rule.Pattern, "attack_type": rule.AttackType, "severity": rule.Severity,
	})
	c.Redirect(http.StatusSeeOther, "/panel/aturan")
}

func (r *Rules) Toggle(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	var cur models.Rule
	if err := r.DB.First(&cur, id).Error; err != nil {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	r.DB.Model(&cur).Update("active", !cur.Active)
	c.Redirect(http.StatusSeeOther, "/panel/aturan")
}

func (r *Rules) Delete(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.String(http.StatusNotFound, "tidak ditemukan")
		return
	}
	r.DB.Delete(&models.Rule{}, id)
	c.Redirect(http.StatusSeeOther, "/panel/aturan")
}

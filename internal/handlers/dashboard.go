package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

var AttackTypes = []string{"Brute force", "Pencurian secret", "Pemindaian admin", "SQL Injection", "XSS", "Path traversal", "Kebocoran repo", "Tidak dikenal"}

const perPage = 25

var likeEsc = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type Dashboard struct{ DB *gorm.DB }

func (d *Dashboard) Index(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	withTest := c.Query("uji") == "1"
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title": "Dashboard", "Username": u.Username, "Role": u.Role,
		"S": services.GetStats(d.DB, withTest), "WithTest": withTest,
	})
}

func (d *Dashboard) Logs(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	withTest := c.Query("uji") == "1"

	// Semua parameter dari URL divalidasi sebelum dipakai.
	typ := c.Query("jenis")
	valid := false
	for _, t := range AttackTypes {
		if t == typ {
			valid = true
		}
	}
	if !valid {
		typ = ""
	}
	q := []rune(strings.TrimSpace(c.Query("q")))
	if len(q) > 60 {
		q = q[:60]
	}
	search := string(q)
	page, _ := strconv.Atoi(c.DefaultQuery("hal", "1"))
	if page < 1 {
		page = 1
	}
	if page > 10000 {
		page = 10000
	}

	tx := d.DB.Model(&models.Event{})
	if !withTest {
		tx = tx.Where("is_test = ?", false)
	}
	if typ != "" {
		tx = tx.Where("attack_type = ?", typ)
	}
	if search != "" {
		like := "%" + likeEsc.Replace(search) + "%"
		tx = tx.Where("path LIKE ? OR ip LIKE ? OR user_agent LIKE ?", like, like, like)
	}
	tx = tx.Session(&gorm.Session{})

	var total int64
	tx.Count(&total)
	var events []models.Event
	tx.Order("id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&events)

	link := func(p int) string {
		v := url.Values{}
		v.Set("hal", strconv.Itoa(p))
		if typ != "" {
			v.Set("jenis", typ)
		}
		if search != "" {
			v.Set("q", search)
		}
		if withTest {
			v.Set("uji", "1")
		}
		return "/panel/log?" + v.Encode()
	}

	data := gin.H{
		"Title": "Log serangan", "Username": u.Username, "Role": u.Role,
		"Events": events, "Total": total, "Types": AttackTypes,
		"Typ": typ, "Q": search, "WithTest": withTest, "Page": page,
	}
	if page > 1 {
		data["Prev"] = link(page - 1)
	}
	if int64(page*perPage) < total {
		data["Next"] = link(page + 1)
	}
	c.HTML(http.StatusOK, "logs.html", data)
}

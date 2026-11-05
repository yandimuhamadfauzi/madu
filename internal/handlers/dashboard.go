package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

var AttackTypes = []string{"Brute force", "Pencurian secret", "Pemindaian admin", "SQL Injection", "XSS", "Path traversal", "Kebocoran repo", "Eksploitasi", "Pemindaian otomatis", "Tidak dikenal"}

const perPage = 25

type Dashboard struct{ DB *gorm.DB }

func (d *Dashboard) Index(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	withTest := c.Query("uji") == "1"
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title": "Dashboard", "CSRF": c.GetString("csrf"), "Username": u.Username, "Role": u.Role,
		"S": services.GetStats(d.DB, withTest), "WithTest": withTest,
	})
}

func (d *Dashboard) Logs(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	f := parseFilter(c)

	page, _ := strconv.Atoi(c.DefaultQuery("hal", "1"))
	if page < 1 {
		page = 1
	}
	if page > 10000 {
		page = 10000
	}

	tx := f.apply(d.DB.Model(&models.Event{})).Session(&gorm.Session{})
	var total int64
	tx.Count(&total)
	var events []models.Event
	tx.Order("id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&events)

	now := time.Now()
	day := func(n int) string { return now.AddDate(0, 0, -n).Format("2006-01-02") }
	rentang := func(n int) string {
		return "/panel/log?" + f.query(map[string]string{"dari": day(n), "sampai": day(0)})
	}

	data := gin.H{
		"Title": "Log serangan", "CSRF": c.GetString("csrf"), "Username": u.Username, "Role": u.Role,
		"Events": events, "Total": total, "Types": AttackTypes, "Page": page,
		"Typ": f.Typ, "Q": f.Q, "WithTest": f.WithTest, "From": f.From, "To": f.To, "Err": f.Err,
		"CSVURL":    "/panel/log.csv?" + f.query(nil),
		"LinkToday": rentang(0), "Link7": rentang(6), "Link30": rentang(29),
		"LinkAll": "/panel/log?" + f.query(nil, "dari", "sampai"),
	}
	if page > 1 {
		data["Prev"] = "/panel/log?" + f.query(map[string]string{"hal": strconv.Itoa(page - 1)})
	}
	if int64(page*perPage) < total {
		data["Next"] = "/panel/log?" + f.query(map[string]string{"hal": strconv.Itoa(page + 1)})
	}
	c.HTML(http.StatusOK, "logs.html", data)
}

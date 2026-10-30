package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9]{3,20}$`)

// Hash palsu: dipakai agar waktu respons sama baik username ada maupun tidak.
var dummyHash, _ = services.HashPassword("kata-sandi-palsu-untuk-menyamakan-waktu")

const msgLoginGagal = "Username atau password salah."

type Auth struct {
	DB           *gorm.DB
	SecureCookie bool
	IPLimit      *services.FailLimiter
	UserLimit    *services.FailLimiter
}

func (a *Auth) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (a *Auth) tooMany(c *gin.Context, d time.Duration) {
	c.Header("Retry-After", strconv.Itoa(int(d.Seconds())+1))
	c.HTML(http.StatusTooManyRequests, "login.html", gin.H{"Error": "Terlalu banyak percobaan. Coba lagi beberapa menit lagi."})
}

func (a *Auth) Login(c *gin.Context) {
	ip := c.ClientIP()
	if d := a.IPLimit.Blocked(ip); d > 0 {
		a.tooMany(c, d)
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")
	fail := func(countUser bool) {
		a.IPLimit.Fail(ip)
		if countUser {
			a.UserLimit.Fail(username)
		}
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": msgLoginGagal})
	}

	if !usernameRe.MatchString(username) || len(password) == 0 || len(password) > 72 {
		fail(false)
		return
	}
	if d := a.UserLimit.Blocked(username); d > 0 {
		a.tooMany(c, d)
		return
	}

	var u models.User
	err := a.DB.Where("username = ?", username).First(&u).Error
	hash := u.PasswordHash
	if err != nil {
		hash = dummyHash
	}
	ok := services.CheckPassword(hash, password)
	if err != nil || !ok {
		fail(true)
		return
	}

	a.UserLimit.Reset(username)
	token, err := services.CreateSession(a.DB, u.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	a.setCookie(c, token, int(services.SessionTTL.Seconds()))
	c.Redirect(http.StatusSeeOther, "/panel")
}

func (a *Auth) Logout(c *gin.Context) {
	if t, err := c.Cookie(services.SessionCookie); err == nil {
		services.DeleteSession(a.DB, t) // sesi dihapus di server, bukan hanya cookie
	}
	a.setCookie(c, "", -1)
	c.Redirect(http.StatusSeeOther, "/masuk")
}

func (a *Auth) setCookie(c *gin.Context, v string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(services.SessionCookie, v, maxAge, "/", "", a.SecureCookie, true) // Secure dan HttpOnly
}

func Panel(c *gin.Context) {
	u := c.MustGet("user").(*models.User)
	c.HTML(http.StatusOK, "panel.html", gin.H{"Username": u.Username, "Role": u.Role})
}

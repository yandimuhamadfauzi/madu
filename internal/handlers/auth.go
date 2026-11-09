package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9]{3,20}$`)

// Hash palsu: dipakai agar waktu respons sama baik username ada maupun tidak.
var dummyHash, _ = services.HashPassword("kata-sandi-palsu-untuk-menyamakan-waktu")

const (
	msgLoginGagal      = "Username atau password salah."
	msgFormKedaluwarsa = "Formulir kedaluwarsa. Silakan coba lagi."
	maxPending         = 50 // batas akun menunggu persetujuan agar tabel tidak bisa dibanjiri
)

type Auth struct {
	DB           *gorm.DB
	SecureCookie bool
	IPLimit      *services.FailLimiter
	UserLimit    *services.FailLimiter
	RegLimit     *services.FailLimiter
}

func (a *Auth) renderLogin(c *gin.Context, status int, msg, info string) {
	c.HTML(status, "login.html", gin.H{"Error": msg, "Info": info, "FT": newFormToken(c, a.SecureCookie)})
}

func (a *Auth) renderRegister(c *gin.Context, status int, msg string) {
	c.HTML(status, "register.html", gin.H{"Error": msg, "FT": newFormToken(c, a.SecureCookie)})
}

func (a *Auth) LoginPage(c *gin.Context) {
	info := ""
	if c.Query("info") == "daftar" {
		info = "Pendaftaran berhasil. Akunmu menunggu persetujuan admin."
	}
	a.renderLogin(c, http.StatusOK, "", info)
}

func (a *Auth) tooMany(c *gin.Context, d time.Duration) {
	c.Header("Retry-After", strconv.Itoa(int(d.Seconds())+1))
	a.renderLogin(c, http.StatusTooManyRequests, "Terlalu banyak percobaan. Coba lagi beberapa menit lagi.", "")
}

func (a *Auth) Login(c *gin.Context) {
	ip := c.ClientIP()
	if d := a.IPLimit.Blocked(ip); d > 0 {
		a.tooMany(c, d)
		return
	}
	if !checkFormToken(c) {
		a.renderLogin(c, http.StatusForbidden, msgFormKedaluwarsa, "")
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")
	key := strings.ToLower(username) // MySQL tidak membedakan huruf besar/kecil, pembatas juga tidak boleh
	fail := func(countUser bool) {
		a.IPLimit.Fail(ip)
		if countUser {
			a.UserLimit.Fail(key)
		}
		a.renderLogin(c, http.StatusUnauthorized, msgLoginGagal, "")
	}

	if !usernameRe.MatchString(username) || len(password) == 0 || len(password) > 72 {
		fail(false)
		return
	}
	if d := a.UserLimit.Blocked(key); d > 0 {
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

	a.UserLimit.Reset(key)
	// Status akun baru diberitahukan setelah password terbukti benar, jadi tidak membocorkan keberadaan akun.
	if u.Status != "aktif" {
		msg := "Akunmu belum aktif. Hubungi admin."
		if u.Status == "menunggu" {
			msg = "Akunmu menunggu persetujuan admin."
		}
		a.renderLogin(c, http.StatusForbidden, msg, "")
		return
	}

	token, err := services.CreateSession(a.DB, u.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	a.setCookie(c, token, int(services.SessionTTL.Seconds()))
	c.Redirect(http.StatusSeeOther, "/panel")
}

func (a *Auth) RegisterPage(c *gin.Context) { a.renderRegister(c, http.StatusOK, "") }

func (a *Auth) Register(c *gin.Context) {
	ip := c.ClientIP()
	if d := a.RegLimit.Blocked(ip); d > 0 {
		c.Header("Retry-After", strconv.Itoa(int(d.Seconds())+1))
		a.renderRegister(c, http.StatusTooManyRequests, "Terlalu banyak percobaan pendaftaran. Coba lagi nanti.")
		return
	}
	if !checkFormToken(c) {
		a.renderRegister(c, http.StatusForbidden, msgFormKedaluwarsa)
		return
	}
	a.RegLimit.Fail(ip) // setiap percobaan dihitung, berhasil maupun gagal

	username := c.PostForm("username")
	email := c.PostForm("email") // Menangkap input email dari form
	pw, pw2 := c.PostForm("password"), c.PostForm("password2")
	bad := func(msg string) { a.renderRegister(c, http.StatusBadRequest, msg) }

	if !usernameRe.MatchString(username) {
		bad("Username harus 3-20 huruf atau angka.")
		return
	}
	// Validasi email sederhana
	if email == "" || !strings.Contains(email, "@") {
		bad("Email tidak valid atau kosong.")
		return
	}
	if pw != pw2 {
		bad("Konfirmasi password tidak cocok.")
		return
	}
	if err := services.ValidatePassword(pw); err != nil {
		bad(err.Error())
		return
	}
	if strings.Contains(strings.ToLower(pw), strings.ToLower(username)) {
		bad("Password tidak boleh memuat username.")
		return
	}

	var pending int64
	a.DB.Model(&models.User{}).Where("status = ?", "menunggu").Count(&pending)
	if pending >= maxPending {
		bad("Pendaftaran sementara penuh. Hubungi admin.")
		return
	}
	var n int64
	a.DB.Model(&models.User{}).Where("username = ?", username).Count(&n)
	if n > 0 {
		bad("Username sudah dipakai.")
		return
	}
	// Periksa email duplikat
	var nEmail int64
	a.DB.Model(&models.User{}).Where("email = ?", email).Count(&nEmail)
	if nEmail > 0 {
		bad("Email sudah terdaftar.")
		return
	}

	h, err := services.HashPassword(pw)
	if err != nil {
		c.String(http.StatusInternalServerError, "terjadi kesalahan")
		return
	}
	// Akun baru: peran analis (hak paling kecil) dan menunggu persetujuan admin.
	newUser := models.User{Username: username, Email: email, PasswordHash: h, Role: "analis", Status: "menunggu"}
	if err := a.DB.Create(&newUser).Error; err != nil {
		bad("Username sudah dipakai.")
		return
	}

	// --- TRIGGER NOTIFIKASI KE ADMIN ---
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail != "" {
		subject := "🚨 Pendaftaran User Baru - Madu Honeypot"
		body := fmt.Sprintf("Halo Admin,\n\nAda pengguna baru yang mendaftar dan menunggu persetujuan Anda.\n\nUsername: %s\nEmail: %s\nRole: %s\n\nSilakan login ke dashboard (https://madu.yandimfauzi.my.id/panel) untuk menyetujui.", newUser.Username, newUser.Email, newUser.Role)

		go func() {
			if err := services.SendEmail([]string{adminEmail}, subject, body); err != nil {
				log.Println("Gagal mengirim email notifikasi ke admin:", err)
			}
		}()
	}
	// -----------------------------------

	c.Redirect(http.StatusSeeOther, "/masuk?info=daftar")
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
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/config"
	"github.com/yandimuhamadfauzi/madu/internal/handlers"
	"github.com/yandimuhamadfauzi/madu/internal/middleware"
	"github.com/yandimuhamadfauzi/madu/internal/models"
	"github.com/yandimuhamadfauzi/madu/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Rule{}, &models.Event{}, &models.Session{}, &models.Blocklist{}, &models.AuditLog{}); err != nil {
		log.Fatal("gagal migrasi: ", err)
	}
	services.SeedUsers(db)
	services.SeedRules(db)

	retention := 90
	if v, err := strconv.Atoi(os.Getenv("EVENT_RETENTION_DAYS")); err == nil && v >= 1 {
		retention = v
	}
	go func() { // tiap jam: bersihkan sesi kedaluwarsa dan log yang melewati masa simpan
		for {
			services.PurgeExpiredSessions(db)
			services.PurgeOldEvents(db, retention)
			time.Sleep(time.Hour)
		}
	}()

	r := gin.Default()
	clientIPHeader := os.Getenv("CLIENT_IP_HEADER") // kosong = pakai alamat koneksi langsung
	if clientIPHeader != "" {
		r.TrustedPlatform = clientIPHeader // aman hanya jika platform hosting menimpa header ini
	}
	_ = r.SetTrustedProxies(nil)
	r.Use(middleware.SanitizeClientIP(clientIPHeader))
	r.Use(middleware.SecurityHeaders(os.Getenv("COOKIE_SECURE") != "false"))
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	auth := &handlers.Auth{
		DB:           db,
		SecureCookie: os.Getenv("COOKIE_SECURE") != "false",
		IPLimit:      services.NewFailLimiter(10, 15*time.Minute, 15*time.Minute),
		UserLimit:    services.NewFailLimiter(5, 15*time.Minute, 15*time.Minute),
	}
	hp := &handlers.Honeypot{DB: db, Limiter: services.NewFailLimiter(60, time.Minute, 10*time.Minute)}
	dash := &handlers.Dashboard{DB: db}
	rules := &handlers.Rules{DB: db}
	blocks := &handlers.Blocks{DB: db}
	export := &handlers.Export{DB: db}
	audit := &handlers.Audit{DB: db}

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/masuk") })
	r.HEAD("/", func(c *gin.Context) { c.Status(http.StatusSeeOther) }) // HEAD yang sah tidak dianggap serangan
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.HEAD("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/masuk", auth.LoginPage)
	r.HEAD("/masuk", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/masuk", middleware.AuditLogin(db), auth.Login)

	requireAuth := middleware.RequireAuth(db)
	csrf := middleware.RequireCSRF()
	admin := middleware.RequireRole("admin")

	r.POST("/keluar", requireAuth, middleware.Audit(db, "logout"), csrf, auth.Logout)
	r.GET("/panel", requireAuth, dash.Index)
	r.GET("/panel/log", requireAuth, dash.Logs)
	r.GET("/panel/log.csv", requireAuth, export.CSV)
	r.GET("/panel/admin", requireAuth, admin, handlers.AdminArea)
	r.GET("/panel/audit", requireAuth, admin, audit.List)

	r.GET("/panel/aturan", requireAuth, rules.List)
	r.POST("/panel/aturan", requireAuth, middleware.Audit(db, "aturan.tambah"), admin, csrf, rules.Create)
	r.POST("/panel/aturan/:id", requireAuth, middleware.Audit(db, "aturan.ubah"), admin, csrf, rules.Update)
	r.POST("/panel/aturan/:id/aktif", requireAuth, middleware.Audit(db, "aturan.aktif"), admin, csrf, rules.Toggle)
	r.POST("/panel/aturan/:id/hapus", requireAuth, middleware.Audit(db, "aturan.hapus"), admin, csrf, rules.Delete)

	r.GET("/panel/blokir", requireAuth, blocks.List)
	r.POST("/panel/blokir", requireAuth, middleware.Audit(db, "blokir.tambah"), admin, csrf, blocks.Add)
	r.POST("/panel/blokir/:id/hapus", requireAuth, middleware.Audit(db, "blokir.hapus"), admin, csrf, blocks.Delete)

	r.NoRoute(hp.Handle) // semua jalur yang tidak ada di aplikasi asli = umpan

	log.Fatal(r.Run(":" + cfg.Port))
}

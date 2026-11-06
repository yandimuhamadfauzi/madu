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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("gagal membaca koneksi database: ", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(&models.User{}, &models.Rule{}, &models.Event{}, &models.Session{}, &models.Blocklist{}, &models.AuditLog{}, &models.Setting{}); err != nil {
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
			services.PurgeExpiredBlocks(db)
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
	r.Use(middleware.LimitBody(64 << 10))
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	auth := &handlers.Auth{
		DB:           db,
		SecureCookie: os.Getenv("COOKIE_SECURE") != "false",
		IPLimit:      services.NewFailLimiter(10, 15*time.Minute, 15*time.Minute),
		UserLimit:    services.NewFailLimiter(5, 15*time.Minute, 15*time.Minute),
		RegLimit:     services.NewFailLimiter(5, time.Hour, time.Hour),
	}
	autoBlocker := services.NewAutoBlocker(services.AutoBlockConfigFromEnv())
	services.LoadAutoBlockSetting(db, autoBlocker)
	log.Println("blokir otomatis aktif:", autoBlocker.Enabled())
	hp := &handlers.Honeypot{DB: db, Limiter: services.NewFailLimiter(60, time.Minute, 10*time.Minute), Global: services.NewFailLimiter(600, time.Minute, time.Minute), AutoBlock: autoBlocker}
	dash := &handlers.Dashboard{DB: db}
	rules := &handlers.Rules{DB: db}
	blocks := &handlers.Blocks{DB: db, Auto: autoBlocker}
	export := &handlers.Export{DB: db}
	audit := &handlers.Audit{DB: db}
	users := &handlers.Users{DB: db}

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/masuk") })
	r.HEAD("/", func(c *gin.Context) { c.Status(http.StatusSeeOther) }) // HEAD yang sah tidak dianggap serangan
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.HEAD("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/masuk", auth.LoginPage)
	r.HEAD("/masuk", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/masuk", middleware.AuditLogin(db), auth.Login)
	r.GET("/daftar", auth.RegisterPage)
	r.HEAD("/daftar", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/daftar", middleware.AuditRegister(db), auth.Register)

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
	r.POST("/panel/aturan", requireAuth, middleware.Audit(db, "aturan.tambah"), middleware.CacheBust(), admin, csrf, rules.Create)
	r.POST("/panel/aturan/:id", requireAuth, middleware.Audit(db, "aturan.ubah"), middleware.CacheBust(), admin, csrf, rules.Update)
	r.POST("/panel/aturan/:id/aktif", requireAuth, middleware.Audit(db, "aturan.aktif"), middleware.CacheBust(), admin, csrf, rules.Toggle)
	r.POST("/panel/aturan/:id/hapus", requireAuth, middleware.Audit(db, "aturan.hapus"), middleware.CacheBust(), admin, csrf, rules.Delete)

	r.GET("/panel/blokir", requireAuth, blocks.List)
	r.POST("/panel/blokir", requireAuth, middleware.Audit(db, "blokir.tambah"), middleware.CacheBust(), admin, csrf, blocks.Add)
	r.POST("/panel/blokir/:id/hapus", requireAuth, middleware.Audit(db, "blokir.hapus"), middleware.CacheBust(), admin, csrf, blocks.Delete)
	r.POST("/panel/blokir/otomatis/aktif", requireAuth, middleware.Audit(db, "autoblock.aktifkan"), admin, csrf, blocks.AutoOn)
	r.POST("/panel/blokir/otomatis/nonaktif", requireAuth, middleware.Audit(db, "autoblock.nonaktifkan"), admin, csrf, blocks.AutoOff)
	r.POST("/panel/blokir/otomatis/hapus", requireAuth, middleware.Audit(db, "autoblock.lepas-semua"), middleware.CacheBust(), admin, csrf, blocks.AutoFlush)

	r.GET("/panel/pengguna", requireAuth, admin, users.List)
	r.POST("/panel/pengguna/:id/setujui", requireAuth, middleware.Audit(db, "pengguna.setujui"), admin, csrf, users.Approve)
	r.POST("/panel/pengguna/:id/nonaktifkan", requireAuth, middleware.Audit(db, "pengguna.nonaktifkan"), admin, csrf, users.Deactivate)
	r.POST("/panel/pengguna/:id/aktifkan", requireAuth, middleware.Audit(db, "pengguna.aktifkan"), admin, csrf, users.Activate)
	r.POST("/panel/pengguna/:id/peran", requireAuth, middleware.Audit(db, "pengguna.peran"), admin, csrf, users.SetRole)
	r.POST("/panel/pengguna/:id/hapus", requireAuth, middleware.Audit(db, "pengguna.hapus"), admin, csrf, users.Delete)

	r.NoRoute(hp.Handle) // semua jalur yang tidak ada di aplikasi asli = umpan

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Println("server berjalan di port " + cfg.Port)
	log.Fatal(srv.ListenAndServe())
}

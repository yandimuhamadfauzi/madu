package main

import (
	"log"
	"net/http"
	"os"
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
	if err := db.AutoMigrate(&models.User{}, &models.Rule{}, &models.Event{}, &models.Session{}, &models.Blocklist{}); err != nil {
		log.Fatal("gagal migrasi: ", err)
	}
	services.SeedUsers(db)
	services.SeedRules(db)

	go func() { // bersihkan sesi kedaluwarsa tiap jam
		for {
			services.PurgeExpiredSessions(db)
			time.Sleep(time.Hour)
		}
	}()

	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // lokal: jangan percaya header proxy. Diatur ulang saat deploy
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

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/masuk") })
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/masuk", auth.LoginPage)
	r.POST("/masuk", auth.Login)

	requireAuth := middleware.RequireAuth(db)
	csrf := middleware.RequireCSRF()
	admin := middleware.RequireRole("admin")

	r.POST("/keluar", requireAuth, csrf, auth.Logout)
	r.GET("/panel", requireAuth, dash.Index)
	r.GET("/panel/log", requireAuth, dash.Logs)
	r.GET("/panel/log.csv", requireAuth, (&handlers.Export{DB: db}).CSV)
	r.GET("/panel/admin", requireAuth, admin, handlers.AdminArea)

	r.GET("/panel/aturan", requireAuth, rules.List)
	r.POST("/panel/aturan", requireAuth, admin, csrf, rules.Create)
	r.POST("/panel/aturan/:id", requireAuth, admin, csrf, rules.Update)
	r.POST("/panel/aturan/:id/aktif", requireAuth, admin, csrf, rules.Toggle)
	r.POST("/panel/aturan/:id/hapus", requireAuth, admin, csrf, rules.Delete)

	r.GET("/panel/blokir", requireAuth, blocks.List)
	r.POST("/panel/blokir", requireAuth, admin, csrf, blocks.Add)
	r.POST("/panel/blokir/:id/hapus", requireAuth, admin, csrf, blocks.Delete)

	r.NoRoute(hp.Handle) // semua jalur yang tidak ada di aplikasi asli = umpan

	log.Fatal(r.Run(":" + cfg.Port))
}

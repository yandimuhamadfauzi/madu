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

	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // lokal: jangan percaya header proxy. Diatur ulang saat deploy
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	auth := &handlers.Auth{
		DB:           db,
		SecureCookie: os.Getenv("COOKIE_SECURE") != "false",
		IPLimit:      services.NewFailLimiter(10, 15*time.Minute, 15*time.Minute),
		UserLimit:    services.NewFailLimiter(5, 15*time.Minute, 15*time.Minute),
	}
	hp := &handlers.Honeypot{DB: db, Limiter: services.NewFailLimiter(60, time.Minute, 10*time.Minute)}

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/masuk") })
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/masuk", auth.LoginPage)
	r.POST("/masuk", auth.Login)
	r.POST("/keluar", auth.Logout)

	requireAuth := middleware.RequireAuth(db)
	r.GET("/panel", requireAuth, handlers.Panel)
	r.GET("/panel/admin", requireAuth, middleware.RequireRole("admin"), handlers.AdminArea)

	r.NoRoute(hp.Handle) // semua jalur yang tidak ada di aplikasi asli = umpan

	log.Fatal(r.Run(":" + cfg.Port))
}

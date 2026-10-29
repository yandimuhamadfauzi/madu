package main

import (
	"log"
	"net/http"
	"os"

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
	if err := db.AutoMigrate(&models.User{}, &models.Rule{}, &models.Event{}, &models.Session{}); err != nil {
		log.Fatal("gagal migrasi: ", err)
	}
	services.SeedAdmin(db)

	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // lokal: jangan percaya header proxy. Diatur ulang saat deploy
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	auth := &handlers.Auth{DB: db, SecureCookie: os.Getenv("COOKIE_SECURE") != "false"}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/masuk", auth.LoginPage)
	r.POST("/masuk", auth.Login)
	r.POST("/keluar", auth.Logout)
	r.GET("/panel", middleware.RequireAuth(db), handlers.Panel)

	log.Fatal(r.Run(":" + cfg.Port))
}

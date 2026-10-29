package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/config"
	"github.com/yandimuhamadfauzi/madu/internal/models"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Rule{}, &models.Event{}); err != nil {
		log.Fatal("gagal migrasi: ", err)
	}

	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // lokal: jangan percaya header proxy. Diatur ulang saat deploy

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Fatal(r.Run(":" + cfg.Port))
}

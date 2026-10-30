package services

import (
	"log"
	"os"
	"regexp"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9]{3,20}$`)

// ensureUser membuat user dari environment variable bila username belum ada.
func ensureUser(db *gorm.DB, uEnv, pEnv, role string) {
	u, p := os.Getenv(uEnv), os.Getenv(pEnv)
	if u == "" || p == "" {
		return
	}
	if !usernameRe.MatchString(u) {
		log.Fatal(uEnv, " harus 3-20 huruf atau angka")
	}
	if err := ValidatePassword(p); err != nil {
		log.Fatal(pEnv, " ditolak: ", err)
	}
	var n int64
	db.Model(&models.User{}).Where("username = ?", u).Count(&n)
	if n > 0 {
		return
	}
	h, err := HashPassword(p)
	if err != nil {
		log.Fatal("gagal membuat hash: ", err)
	}
	if err := db.Create(&models.User{Username: u, PasswordHash: h, Role: role}).Error; err != nil {
		log.Fatal("gagal membuat user: ", err)
	}
	log.Println("user dibuat:", u, "peran:", role)
}

func SeedUsers(db *gorm.DB) {
	ensureUser(db, "ADMIN_USERNAME", "ADMIN_PASSWORD", "admin")
	ensureUser(db, "ANALIS_USERNAME", "ANALIS_PASSWORD", "analis")
}

package services

import (
	"log"
	"os"
	"regexp"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9]{3,20}$`)

// SeedAdmin membuat admin pertama dari environment variable, hanya jika belum ada user.
func SeedAdmin(db *gorm.DB) {
	var n int64
	db.Model(&models.User{}).Count(&n)
	if n > 0 {
		return
	}
	u, p := os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD")
	if u == "" || p == "" {
		log.Println("belum ada user: isi ADMIN_USERNAME dan ADMIN_PASSWORD di .env untuk membuat admin pertama")
		return
	}
	if !usernameRe.MatchString(u) {
		log.Fatal("ADMIN_USERNAME harus 3-20 huruf atau angka")
	}
	if err := ValidatePassword(p); err != nil {
		log.Fatal("ADMIN_PASSWORD ditolak: ", err)
	}
	h, err := HashPassword(p)
	if err != nil {
		log.Fatal("gagal membuat hash: ", err)
	}
	if err := db.Create(&models.User{Username: u, PasswordHash: h, Role: "admin"}).Error; err != nil {
		log.Fatal("gagal membuat admin: ", err)
	}
	log.Println("admin pertama dibuat:", u)
}

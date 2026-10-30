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

func SeedRules(db *gorm.DB) {
	var n int64
	db.Model(&models.Rule{}).Count(&n)
	if n > 0 {
		return
	}
	defaults := []models.Rule{
		{Name: "WordPress login", Pattern: "wp-login", AttackType: "Brute force", Severity: 2, Active: true},
		{Name: "WordPress XML-RPC", Pattern: "xmlrpc", AttackType: "Brute force", Severity: 2, Active: true},
		{Name: "File konfigurasi", Pattern: ".env", AttackType: "Pencurian secret", Severity: 3, Active: true},
		{Name: "Panel database", Pattern: "phpmyadmin", AttackType: "Pemindaian admin", Severity: 2, Active: true},
		{Name: "Halaman admin", Pattern: "/admin", AttackType: "Pemindaian admin", Severity: 1, Active: true},
		{Name: "Pola SQLi", Pattern: "' or ", AttackType: "SQL Injection", Severity: 3, Active: true},
		{Name: "UNION SELECT", Pattern: "union select", AttackType: "SQL Injection", Severity: 3, Active: true},
		{Name: "Tag script", Pattern: "<script", AttackType: "XSS", Severity: 3, Active: true},
		{Name: "Path traversal", Pattern: "../", AttackType: "Path traversal", Severity: 3, Active: true},
		{Name: "Repositori Git", Pattern: ".git/", AttackType: "Kebocoran repo", Severity: 3, Active: true},
	}
	db.Create(&defaults)
}

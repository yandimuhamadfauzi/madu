package services

import (
	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

// SeedRules mengisi aturan awal hanya jika tabel masih kosong.
// Pola dicocokkan terhadap path+query (dan user-agent) yang sudah di-decode dan diubah ke huruf kecil.
func SeedRules(db *gorm.DB) {
	var n int64
	db.Model(&models.Rule{}).Count(&n)
	if n > 0 {
		return
	}
	r := func(name, pattern, typ string, sev int) models.Rule {
		return models.Rule{Name: name, Pattern: pattern, AttackType: typ, Severity: sev, Active: true}
	}
	rules := []models.Rule{
		// Pencurian secret dan file konfigurasi
		r("File .env", ".env", "Pencurian secret", 3),
		r("Kredensial AWS", "/.aws/", "Pencurian secret", 3),
		r("Kunci SSH", "/.ssh/", "Pencurian secret", 3),
		r("Kunci privat", "id_rsa", "Pencurian secret", 3),
		r("Berkas htpasswd", ".htpasswd", "Pencurian secret", 3),
		r("Konfigurasi WordPress", "wp-config", "Pencurian secret", 3),
		r("Dump database", ".sql", "Pencurian secret", 2),
		// Kebocoran repositori
		r("Repositori Git", ".git/", "Kebocoran repo", 3),
		r("Repositori SVN", "/.svn/", "Kebocoran repo", 3),
		// Path traversal
		r("Path traversal", "../", "Path traversal", 3),
		r("Path traversal (backslash)", "..\\", "Path traversal", 3),
		r("File sistem Linux", "/etc/passwd", "Path traversal", 3),
		r("File sistem Windows", "win.ini", "Path traversal", 3),
		// SQL Injection
		r("Pola SQLi OR", "' or ", "SQL Injection", 3),
		r("UNION SELECT", "union select", "SQL Injection", 3),
		r("SQLi berbasis waktu", "sleep(", "SQL Injection", 3),
		r("Skema database", "information_schema", "SQL Injection", 3),
		r("Tautologi", "or 1=1", "SQL Injection", 3),
		// XSS
		r("Tag script", "<script", "XSS", 3),
		r("URI javascript", "javascript:", "XSS", 3),
		r("Handler onerror", "onerror=", "XSS", 3),
		r("Tag svg", "<svg", "XSS", 3),
		// Eksploitasi
		r("Log4Shell", "${jndi:", "Eksploitasi", 3),
		r("PHPUnit RCE", "phpunit", "Eksploitasi", 3),
		r("Shell sistem", "/bin/sh", "Eksploitasi", 3),
		r("Shellshock", "() {", "Eksploitasi", 3),
		r("Fungsi PHP berbahaya", "base64_decode", "Eksploitasi", 3),
		// Brute force dan login palsu
		r("WordPress login", "wp-login", "Brute force", 2),
		r("WordPress XML-RPC", "xmlrpc", "Brute force", 2),
		r("Halaman login umum", "/login", "Brute force", 2),
		// Pemindaian panel admin
		r("Panel database", "phpmyadmin", "Pemindaian admin", 2),
		r("Adminer", "adminer", "Pemindaian admin", 2),
		r("Spring Actuator", "/actuator", "Pemindaian admin", 2),
		r("Server status", "/server-status", "Pemindaian admin", 2),
		r("Halaman admin", "/admin", "Pemindaian admin", 1),
		r("Dashboard palsu", "/dashboard", "Pemindaian admin", 1),
		// Alat pemindai (dari user-agent)
		r("sqlmap", "sqlmap", "SQL Injection", 3),
		r("Nikto", "nikto", "Pemindaian otomatis", 2),
		r("Nmap", "nmap", "Pemindaian otomatis", 2),
		r("Masscan", "masscan", "Pemindaian otomatis", 2),
		r("zgrab", "zgrab", "Pemindaian otomatis", 2),
		r("Nuclei", "nuclei", "Pemindaian otomatis", 2),
		r("Gobuster", "gobuster", "Pemindaian otomatis", 2),
		r("WPScan", "wpscan", "Pemindaian otomatis", 2),
		r("Hydra", "hydra", "Brute force", 2),
	}
	db.Create(&rules)
}

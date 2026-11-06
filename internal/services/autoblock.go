package services

import (
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// AutoBlockConfig mengatur kapan sebuah IP diblokir otomatis oleh jalur jebakan.
type AutoBlockConfig struct {
	Enabled     bool
	Threshold   int           // jumlah serangan yang memicu blokir
	Window      time.Duration // jendela waktu penghitungan
	Duration    time.Duration // lama blokir
	MinSeverity int           // hanya serangan dengan severity >= ini yang dihitung
}

func DefaultAutoBlockConfig() AutoBlockConfig {
	return AutoBlockConfig{Enabled: true, Threshold: 5, Window: 10 * time.Minute, Duration: time.Hour, MinSeverity: 3}
}

// AutoBlockConfigFromEnv membaca AUTOBLOCK_ENABLED (false untuk mematikan), AUTOBLOCK_THRESHOLD,
// AUTOBLOCK_WINDOW_MINUTES, dan AUTOBLOCK_DURATION_MINUTES. Nilai yang kosong atau tidak valid memakai bawaan.
func AutoBlockConfigFromEnv() AutoBlockConfig {
	c := DefaultAutoBlockConfig()
	if os.Getenv("AUTOBLOCK_ENABLED") == "false" {
		c.Enabled = false
	}
	c.Threshold = envInt("AUTOBLOCK_THRESHOLD", c.Threshold, 1)
	c.Window = time.Duration(envInt("AUTOBLOCK_WINDOW_MINUTES", int(c.Window/time.Minute), 1)) * time.Minute
	c.Duration = time.Duration(envInt("AUTOBLOCK_DURATION_MINUTES", int(c.Duration/time.Minute), 1)) * time.Minute
	return c
}

// envInt membaca bilangan bulat dari environment variable; nilai kosong, bukan angka, atau di bawah min memakai def.
func envInt(key string, def, min int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil || v < min {
		return def
	}
	return v
}

// AutoBlocker menghitung serangan berat per IP dan memberi tahu kapan sebuah IP harus diblokir.
// Hitungan disimpan di memori (hilang saat restart), sama seperti pembatas login.
//
// Fitur ini punya dua lapis saklar: AUTOBLOCK_ENABLED=false di environment mematikannya dan mengunci
// (tidak bisa dinyalakan dari web), sedangkan di bawah itu admin bisa menyalakan/mematikannya dari halaman Blokir.
type AutoBlocker struct {
	cfg    AutoBlockConfig
	hits   *FailLimiter
	locked bool        // dimatikan lewat environment
	on     atomic.Bool // pilihan admin saat runtime
}

func NewAutoBlocker(cfg AutoBlockConfig) *AutoBlocker {
	if cfg.Threshold < 1 {
		cfg.Threshold = 1
	}
	a := &AutoBlocker{cfg: cfg, hits: NewFailLimiter(cfg.Threshold, cfg.Window, cfg.Window), locked: !cfg.Enabled}
	a.on.Store(cfg.Enabled)
	return a
}

func (a *AutoBlocker) Config() AutoBlockConfig { return a.cfg }

// Enabled melaporkan apakah blokir otomatis sedang berjalan.
func (a *AutoBlocker) Enabled() bool { return a != nil && !a.locked && a.on.Load() }

// Locked true bila fitur dimatikan lewat environment dan tidak bisa diubah dari web.
func (a *AutoBlocker) Locked() bool { return a != nil && a.locked }

// SetEnabled menyalakan/mematikan blokir otomatis saat runtime. Tidak berpengaruh bila Locked.
func (a *AutoBlocker) SetEnabled(v bool) {
	if a != nil && !a.locked {
		a.on.Store(v)
	}
}

// Record mencatat satu serangan dan mengembalikan true tepat sekali saat ambang tercapai.
// Alamat lokal, privat, dan sejenisnya tidak pernah dihitung: bila CLIENT_IP_HEADER belum diatur,
// IP yang terbaca adalah IP proxy hosting, dan memblokirnya akan memblokir semua pengunjung.
func (a *AutoBlocker) Record(ip string, severity int) bool {
	if !a.Enabled() || severity < a.cfg.MinSeverity || !blockableIP(ip) {
		return false
	}
	if a.hits.Blocked(ip) > 0 {
		return false
	}
	return a.hits.Fail(ip)
}

var cgnat = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// blockableIP hanya menerima alamat IP publik yang valid.
func blockableIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || cgnat.Contains(ip) {
		return false
	}
	return true
}

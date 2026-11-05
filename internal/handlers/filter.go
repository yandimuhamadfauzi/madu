package handlers

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var likeEsc = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// eventFilter memuat filter log serangan yang sudah divalidasi. Dipakai oleh halaman log dan ekspor CSV.
type eventFilter struct {
	Typ        string
	Q          string
	WithTest   bool
	From, To   string // YYYY-MM-DD yang sudah divalidasi (kosong = tidak dipakai)
	fromT, toT time.Time
	Err        string
}

func parseFilter(c *gin.Context) eventFilter {
	f := eventFilter{WithTest: c.Query("uji") == "1"}

	// Jenis hanya boleh salah satu nilai yang diizinkan.
	typ := c.Query("jenis")
	for _, t := range AttackTypes {
		if t == typ {
			f.Typ = typ
		}
	}
	q := []rune(strings.TrimSpace(c.Query("q")))
	if len(q) > 60 {
		q = q[:60]
	}
	f.Q = string(q)

	parse := func(s string) (time.Time, bool) {
		if s == "" {
			return time.Time{}, false
		}
		t, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil || t.Year() < 2000 || t.Year() > 2100 {
			f.Err = "Format tanggal tidak valid (gunakan tahun-bulan-tanggal)."
			return time.Time{}, false
		}
		return t, true
	}
	from, okF := parse(c.Query("dari"))
	to, okT := parse(c.Query("sampai"))
	if okF && okT && from.After(to) {
		f.Err = "Tanggal awal tidak boleh setelah tanggal akhir."
		okF, okT = false, false
	}
	if okF {
		f.From, f.fromT = from.Format("2006-01-02"), from
	}
	if okT {
		f.To, f.toT = to.Format("2006-01-02"), to.AddDate(0, 0, 1) // batas atas eksklusif: seluruh hari "sampai" ikut
	}
	return f
}

// apply menambahkan kondisi filter ke query. Semua nilai memakai parameter (?).
func (f eventFilter) apply(tx *gorm.DB) *gorm.DB {
	if !f.WithTest {
		tx = tx.Where("is_test = ?", false)
	}
	if f.Typ != "" {
		tx = tx.Where("attack_type = ?", f.Typ)
	}
	if f.Q != "" {
		like := "%" + likeEsc.Replace(f.Q) + "%"
		tx = tx.Where("path LIKE ? OR ip LIKE ? OR user_agent LIKE ?", like, like, like)
	}
	if !f.fromT.IsZero() {
		tx = tx.Where("created_at >= ?", f.fromT)
	}
	if !f.toT.IsZero() {
		tx = tx.Where("created_at < ?", f.toT)
	}
	return tx
}

func (f eventFilter) values() url.Values {
	v := url.Values{}
	if f.Typ != "" {
		v.Set("jenis", f.Typ)
	}
	if f.Q != "" {
		v.Set("q", f.Q)
	}
	if f.WithTest {
		v.Set("uji", "1")
	}
	if f.From != "" {
		v.Set("dari", f.From)
	}
	if f.To != "" {
		v.Set("sampai", f.To)
	}
	return v
}

// query menyusun query string dari filter, dengan nilai tambahan (set) dan kunci yang dibuang (drop).
func (f eventFilter) query(set map[string]string, drop ...string) string {
	v := f.values()
	for _, k := range drop {
		v.Del(k)
	}
	for k, val := range set {
		v.Set(k, val)
	}
	return v.Encode()
}

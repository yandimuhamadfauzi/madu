package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"gorm.io/gorm"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

const (
	SessionCookie = "sid"
	SessionTTL    = 8 * time.Hour
)

// Yang disimpan di database hanya hash token, jadi kebocoran database tidak membocorkan sesi aktif.
func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

func randomToken() (string, error) {
	b := make([]byte, 32) // 256 bit acak dari crypto/rand
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func CreateSession(db *gorm.DB, userID uint) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	csrf, err := randomToken()
	if err != nil {
		return "", err
	}
	s := models.Session{UserID: userID, TokenHash: hashToken(token), CSRFToken: csrf, ExpiresAt: time.Now().Add(SessionTTL)}
	if err := db.Create(&s).Error; err != nil {
		return "", err
	}
	return token, nil
}

// Sesi tanpa token CSRF (sesi lama) dianggap tidak valid.
func SessionFromToken(db *gorm.DB, token string) (*models.User, *models.Session, error) {
	var s models.Session
	if err := db.Where("token_hash = ? AND expires_at > ? AND csrf_token <> ?", hashToken(token), time.Now(), "").First(&s).Error; err != nil {
		return nil, nil, err
	}
	var u models.User
	if err := db.First(&u, s.UserID).Error; err != nil {
		return nil, nil, err
	}
	return &u, &s, nil
}

func DeleteSession(db *gorm.DB, token string) {
	db.Where("token_hash = ?", hashToken(token)).Delete(&models.Session{})
}

func PurgeExpiredSessions(db *gorm.DB) {
	db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
}

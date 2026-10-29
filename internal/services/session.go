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

func CreateSession(db *gorm.DB, userID uint) (string, error) {
	b := make([]byte, 32) // 256 bit acak dari crypto/rand
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	s := models.Session{UserID: userID, TokenHash: hashToken(token), ExpiresAt: time.Now().Add(SessionTTL)}
	if err := db.Create(&s).Error; err != nil {
		return "", err
	}
	return token, nil
}

func UserFromToken(db *gorm.DB, token string) (*models.User, error) {
	var s models.Session
	if err := db.Where("token_hash = ? AND expires_at > ?", hashToken(token), time.Now()).First(&s).Error; err != nil {
		return nil, err
	}
	var u models.User
	if err := db.First(&u, s.UserID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func DeleteSession(db *gorm.DB, token string) {
	db.Where("token_hash = ?", hashToken(token)).Delete(&models.Session{})
}

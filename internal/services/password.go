package services

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var ErrWeakPassword = errors.New("password minimal 8 karakter dengan huruf besar, huruf kecil, dan angka")

// ValidatePassword menolak password lemah. Batas 72 byte karena itu batas bcrypt.
func ValidatePassword(pw string) error {
	if len(pw) < 8 || len(pw) > 72 {
		return ErrWeakPassword
	}
	var up, lo, di bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			up = true
		case unicode.IsLower(r):
			lo = true
		case unicode.IsDigit(r):
			di = true
		}
	}
	if !(up && lo && di) {
		return ErrWeakPassword
	}
	return nil
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

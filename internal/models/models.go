package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:20;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:100;not null"`
	Role         string `gorm:"size:10;not null"` // admin atau analis
	CreatedAt    time.Time
}

type Rule struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"size:40;not null"`
	Pattern    string `gorm:"size:40;not null"`
	AttackType string `gorm:"size:30;not null"`
	Severity   int
	Active     bool `gorm:"default:true"`
	CreatedAt  time.Time
}

type Event struct {
	ID         uint      `gorm:"primaryKey"`
	CreatedAt  time.Time `gorm:"index"`
	IP         string    `gorm:"size:45;index"`
	Method     string    `gorm:"size:10"`
	Path       string    `gorm:"size:500"`
	UserAgent  string    `gorm:"size:300"`
	AttackType string    `gorm:"size:30;index"`
	Severity   int
	Blocked    bool
	IsTest     bool
}

type Session struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

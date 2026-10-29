package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DSN  string
}

func Load() Config {
	_ = godotenv.Load() // lokal: baca .env. Di hosting: nilai datang dari environment variable

	get := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			panic("environment variable belum diisi: " + k)
		}
		return v
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		get("DB_USER"), get("DB_PASSWORD"), get("DB_HOST"), get("DB_PORT"), get("DB_NAME"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port, DSN: dsn}
}

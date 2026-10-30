package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
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

	// Koneksi terenkripsi ke database terkelola.
	// DB_CA_CERT berisi sertifikat CA (format PEM). DB_TLS=true memakai CA sistem (cocok untuk beberapa penyedia).
	tlsParam := ""
	if ca := os.Getenv("DB_CA_CERT"); ca != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(ca)) {
			panic("DB_CA_CERT bukan sertifikat PEM yang valid")
		}
		if err := mysql.RegisterTLSConfig("custom", &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}); err != nil {
			panic(err)
		}
		tlsParam = "&tls=custom"
	} else if os.Getenv("DB_TLS") == "true" {
		tlsParam = "&tls=true"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
		get("DB_USER"), get("DB_PASSWORD"), get("DB_HOST"), get("DB_PORT"), get("DB_NAME"), tlsParam)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port, DSN: dsn}
}

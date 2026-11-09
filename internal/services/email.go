package services

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail mengirimkan email menggunakan konfigurasi SMTP di .env
func SendEmail(to []string, subject string, message string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	if from == "" || password == "" {
		return fmt.Errorf("konfigurasi SMTP belum diatur di .env")
	}

	// Autentikasi ke server SMTP
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Format header yang lebih lengkap agar mengurangi risiko masuk folder Spam
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("\"Madu\" <%s>", from)
	header["To"] = to[0]
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"UTF-8\""

	var msgString string
	for k, v := range header {
		msgString += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	
	// Tambahkan baris kosong (\r\n) sebagai pemisah mutlak antara header dan isi (body) email
	msgString += "\r\n" + message

	// Kirim email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(msgString))
	return err
}
package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends an email using Gmail SMTP
func SendEmail(to, subject, body string) error {
	email := os.Getenv("SMTP_EMAIL")
	fmt.Errorf("email", email)
	password := os.Getenv("SMTP_PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("SMTP_EMAIL or SMTP_PASSWORD not set")
	}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Set up authentication information.
	auth := smtp.PlainAuth("", email, password, smtpHost)

	// Compose the message
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, email, []string{to}, msg)
	if err != nil {
		// Retry once
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, email, []string{to}, msg)
	}
	return err
}

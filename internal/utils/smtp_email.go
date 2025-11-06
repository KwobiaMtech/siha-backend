package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendRealEmail(to, subject, htmlContent, textContent string) error {
	// Gmail SMTP configuration
	smtpHost := getEnvOrDefault("SMTP_HOST", "smtp.gmail.com")
	smtpPortStr := getEnvOrDefault("SMTP_PORT", "587")
	smtpUser := os.Getenv("SMTP_USER") // Gmail address
	smtpPass := os.Getenv("SMTP_PASS") // App password

	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		smtpPort = 587
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textContent)
	m.AddAlternative("text/html", htmlContent)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("📧 Real email sent successfully to: %s", to)
	return nil
}

// Update SendVerificationEmail to try real email first
func SendVerificationEmailReal(email, code string) error {
	subject := "Verify Your Email - SIHA"
	htmlContent := createVerificationEmailHTML(code)
	textContent := createVerificationEmailText(code)

	// Try SMTP first
	err := SendRealEmail(email, subject, htmlContent, textContent)
	if err != nil {
		log.Printf("⚠️ Failed to send real email: %v", err)
		// Fallback to original method
		return SendVerificationEmail(email, code)
	}

	return nil
}

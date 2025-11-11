package utils

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type EmailService struct {
	brevoSender *BrevoEmailSender
	fromEmail   string
	fromName    string
	baseURL     string
}

func NewEmailService() *EmailService {
	brevoAPIKey := os.Getenv("BREVO_API_KEY")
	var brevoSender *BrevoEmailSender
	if brevoAPIKey != "" {
		brevoSender = NewBrevoEmailSender(brevoAPIKey)
	}

	return &EmailService{
		brevoSender: brevoSender,
		fromEmail:   getEnvOrDefault("SENDER_EMAIL", "noreply@siha.com"),
		fromName:    getEnvOrDefault("SENDER_NAME", "SIHA"),
		baseURL:     getEnvOrDefault("BASE_URL", "https://siha.com"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type BrevoEmailSender struct {
	APIKey string
	client *brevo.APIClient
}

func NewBrevoEmailSender(apiKey string) *BrevoEmailSender {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	client := brevo.NewAPIClient(cfg)

	return &BrevoEmailSender{
		APIKey: apiKey,
		client: client,
	}
}

func (b *BrevoEmailSender) SendEmail(to string, toName string, subject string, htmlContent string, textContent string) error {
	ctx := context.Background()

	sender := &brevo.SendSmtpEmailSender{
		Email: getEnvOrDefault("SENDER_EMAIL", "noreply@siha.com"),
		Name:  getEnvOrDefault("SENDER_NAME", "SIHA"),
	}

	recipient := brevo.SendSmtpEmailTo{
		Email: to,
		Name:  toName,
	}

	sendSmtpEmail := &brevo.SendSmtpEmail{
		Sender:      sender,
		To:          []brevo.SendSmtpEmailTo{recipient},
		Subject:     subject,
		HtmlContent: htmlContent,
		TextContent: textContent,
	}

	result, resp, err := b.client.TransactionalEmailsApi.SendTransacEmail(ctx, *sendSmtpEmail)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully. Status: %d, Message ID: %s", resp.StatusCode, result.MessageId)
	return nil
}

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendVerificationEmail(email, code string) error {
	emailService := NewEmailService()
	
	subject := "Verify Your Email - SIHA"
	htmlContent := createVerificationEmailHTML(code)
	textContent := createVerificationEmailText(code)

	// Try SMTP first if configured
	if os.Getenv("EMAIL_PROVIDER") == "smtp" || os.Getenv("BREVO_API_KEY") == "" {
		err := SendRealEmail(email, subject, htmlContent, textContent)
		if err == nil {
			log.Printf("✅ Verification email sent to %s via SMTP", email)
			return nil
		}
		log.Printf("⚠️ SMTP failed: %v, trying Brevo...", err)
	}

	// Fallback to Brevo if available
	if emailService.brevoSender != nil {
		err := emailService.brevoSender.SendEmail(email, "", subject, htmlContent, textContent)
		if err != nil {
			log.Printf("❌ Failed to send email via Brevo: %v", err)
			return fmt.Errorf("failed to send email via both SMTP and Brevo: %w", err)
		}
		log.Printf("✅ Verification email sent to %s via Brevo", email)
		return nil
	}

	return fmt.Errorf("no email service configured - please set SMTP credentials or BREVO_API_KEY")
}

func createVerificationEmailHTML(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification - SIHA</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; box-shadow: 0 0 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #4CAF50 0%%, #45a049 100%%); color: white; padding: 40px 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 32px; font-weight: 300; }
        .header p { margin: 10px 0 0 0; opacity: 0.9; font-size: 16px; }
        .content { padding: 40px 30px; }
        .verification-code { background: linear-gradient(135deg, #4CAF50 0%%, #45a049 100%%); color: white; font-size: 36px; font-weight: bold; text-align: center; padding: 25px; margin: 30px 0; border-radius: 12px; letter-spacing: 8px; }
        .instructions { background-color: #f8f9fa; padding: 25px; border-radius: 8px; border-left: 4px solid #4CAF50; margin: 25px 0; }
        .footer { background-color: #2c3e50; color: white; padding: 25px; text-align: center; }
        .highlight-box { background: linear-gradient(135deg, #4CAF5020 0%%, #45a04920 100%%); padding: 20px; border-radius: 8px; margin: 20px 0; border: 1px solid #4CAF5040; }
        .security-note { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 8px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏥 SIHA</h1>
            <p>Secure Healthcare & Financial Services</p>
        </div>
        
        <div class="content">
            <h2 style="color: #4CAF50; margin-bottom: 20px;">📧 Email Verification Required</h2>
            
            <p>Welcome to SIHA! To complete your account setup and ensure the security of your healthcare and financial data, please verify your email address.</p>
            
            <div class="instructions">
                <h3 style="margin-top: 0; color: #4CAF50;">✅ Verification Instructions:</h3>
                <ol>
                    <li>Copy the 6-digit verification code below</li>
                    <li>Return to the SIHA application</li>
                    <li>Enter the code in the verification field</li>
                    <li>Click "Verify Email" to complete setup</li>
                </ol>
            </div>
            
            <div class="verification-code">
                %s
            </div>
            
            <div class="highlight-box">
                <h4 style="margin-top: 0; color: #4CAF50;">🔒 Why Email Verification?</h4>
                <ul style="margin: 10px 0;">
                    <li>✅ Secure your healthcare data and financial information</li>
                    <li>✅ Enable account recovery and important notifications</li>
                    <li>✅ Comply with healthcare data protection standards</li>
                    <li>✅ Prevent unauthorized access to your account</li>
                </ul>
            </div>
            
            <div class="security-note">
                <h4 style="margin-top: 0; color: #856404;">⚠️ Security Notice:</h4>
                <ul style="margin: 10px 0 0 0;">
                    <li>This code expires in 15 minutes for your security</li>
                    <li>Never share this code with anyone</li>
                    <li>SIHA staff will never ask for your verification code</li>
                    <li>If you didn't request this, please ignore this email</li>
                </ul>
            </div>
            
            <div style="border-left: 4px solid #17a2b8; padding-left: 20px; margin: 25px 0; background-color: #e7f3ff; padding: 15px;">
                <h4 style="margin-top: 0; color: #17a2b8;">📞 Need Help?</h4>
                <p style="margin-bottom: 0;">If you're having trouble with verification:</p>
                <ul style="margin: 10px 0 0 0;">
                    <li>📧 Email: support@siha.com</li>
                    <li>📱 Phone: +233 (0) 123 456 789</li>
                    <li>🌐 Help Center: help.siha.com</li>
                </ul>
            </div>
            
            <p style="margin-top: 30px;">Thank you for choosing SIHA!</p>
            
            <p>Best regards,<br>
            <strong>The SIHA Team</strong><br>
            <em>Your Trusted Healthcare & Financial Partner</em></p>
        </div>
        
        <div class="footer">
            <p><strong>📧 Secure Email Service</strong></p>
            <p>SIHA - Secure Healthcare & Financial Services | Est. 2024</p>
            <p style="margin: 0; opacity: 0.8;">© 2024 SIHA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, code)
}

func createVerificationEmailText(code string) string {
	return fmt.Sprintf(`
SIHA - Email Verification Required

Welcome to SIHA!

Your verification code is: %s

To complete your account setup:
1. Return to the SIHA application
2. Enter this 6-digit code in the verification field
3. Click "Verify Email"

This code expires in 15 minutes for your security.

Why verify your email?
- Secure your healthcare and financial data
- Enable account recovery
- Receive important notifications
- Comply with data protection standards

Security Notice:
- Never share this code with anyone
- SIHA staff will never ask for your verification code
- If you didn't request this, please ignore this email

Need help?
- Email: support@siha.com
- Phone: +233 (0) 123 456 789

Thank you for choosing SIHA!

Best regards,
The SIHA Team

© 2024 SIHA. All rights reserved.
`, code)
}

package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// Service handles email sending
type Service struct {
	config *config.SMTPConfig
}

// NewService creates a new email service
func NewService(cfg *config.SMTPConfig) *Service {
	return &Service{
		config: cfg,
	}
}

// EmailData holds data for email templates
type EmailData struct {
	To      string
	Subject string
	Data    map[string]interface{}
}

// SendPlainEmail sends a plain text email
func (s *Service) SendPlainEmail(to, subject, body string) error {
	// Setup authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose message
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body))

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, msg)
	if err != nil {
		logger.Error("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	return nil
}

// SendHTMLEmail sends an HTML email
func (s *Service) SendHTMLEmail(to, subject, htmlBody string) error {
	// Setup authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose message with HTML
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, htmlBody))

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, msg)
	if err != nil {
		logger.Error("Failed to send HTML email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send HTML email: %w", err)
	}

	logger.Info("HTML email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	return nil
}

// SendTemplateEmail sends an email using a template
func (s *Service) SendTemplateEmail(to, subject, templateName string, data map[string]interface{}) error {
	// Get template
	tmpl, err := s.getTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		logger.Error("Failed to execute template",
			zap.String("template", templateName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Send HTML email
	return s.SendHTMLEmail(to, subject, body.String())
}

// getTemplate returns a template by name
func (s *Service) getTemplate(name string) (*template.Template, error) {
	templates := map[string]string{
		"welcome":        welcomeTemplate,
		"verify_email":   verifyEmailTemplate,
		"reset_password": resetPasswordTemplate,
		"otp":            otpTemplate,
	}

	tmplStr, ok := templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// Email Templates
const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to GO CMS</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4CAF50;">Welcome to GO CMS!</h1>
        <p>Hi {{.Name}},</p>
        <p>Thank you for registering with GO CMS. We're excited to have you on board!</p>
        <p>Your account has been successfully created.</p>
        <p>Best regards,<br>The GO CMS Team</p>
    </div>
</body>
</html>
`

const verifyEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4CAF50;">Verify Your Email Address</h1>
        <p>Hi {{.Name}},</p>
        <p>Please use the following code to verify your email address:</p>
        <div style="background-color: #f4f4f4; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; margin: 20px 0;">
            {{.OTP}}
        </div>
        <p style="color: #666; font-size: 14px;">This code will expire in {{.ExpiryMinutes}} seconds.</p>
        <p>If you didn't request this, please ignore this email.</p>
        <p>Best regards,<br>The GO CMS Team</p>
    </div>
</body>
</html>
`

const resetPasswordTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #FF5722;">Reset Your Password</h1>
        <p>Hi {{.Name}},</p>
        <p>We received a request to reset your password. Use the following code to reset it:</p>
        <div style="background-color: #f4f4f4; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; margin: 20px 0;">
            {{.OTP}}
        </div>
        <p style="color: #666; font-size: 14px;">This code will expire in {{.ExpiryMinutes}} seconds.</p>
        <p>If you didn't request this, please ignore this email and your password will remain unchanged.</p>
        <p>Best regards,<br>The GO CMS Team</p>
    </div>
</body>
</html>
`

const otpTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Verification Code</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2196F3;">Your Verification Code</h1>
        <p>Hi {{.Name}},</p>
        <p>{{.Message}}</p>
        <div style="background-color: #f4f4f4; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; margin: 20px 0;">
            {{.OTP}}
        </div>
        <p style="color: #666; font-size: 14px;">This code will expire in {{.ExpiryMinutes}} seconds.</p>
        <p>If you didn't request this, please ignore this email.</p>
        <p>Best regards,<br>The GO CMS Team</p>
    </div>
</body>
</html>
`

// Helper functions for common email scenarios

// SendOTPEmail sends an OTP verification email
func (s *Service) SendOTPEmail(to, name, otp string, expirySeconds int) error {
	data := map[string]interface{}{
		"Name":          name,
		"OTP":           otp,
		"ExpiryMinutes": expirySeconds,
		"Message":       "Please use the following code to verify your action:",
	}

	return s.SendTemplateEmail(to, "Your Verification Code", "otp", data)
}

// SendVerifyEmailOTP sends email verification OTP
func (s *Service) SendVerifyEmailOTP(to, name, otp string, expirySeconds int) error {
	data := map[string]interface{}{
		"Name":          name,
		"OTP":           otp,
		"ExpiryMinutes": expirySeconds,
	}

	return s.SendTemplateEmail(to, "Verify Your Email Address", "verify_email", data)
}

// SendResetPasswordOTP sends password reset OTP
func (s *Service) SendResetPasswordOTP(to, name, otp string, expirySeconds int) error {
	data := map[string]interface{}{
		"Name":          name,
		"OTP":           otp,
		"ExpiryMinutes": expirySeconds,
	}

	return s.SendTemplateEmail(to, "Reset Your Password", "reset_password", data)
}

// SendWelcomeEmail sends a welcome email
func (s *Service) SendWelcomeEmail(to, name string) error {
	data := map[string]interface{}{
		"Name": name,
	}

	return s.SendTemplateEmail(to, "Welcome to GO CMS", "welcome", data)
}

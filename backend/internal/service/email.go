package service

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
	AppURL       string
}

type EmailService struct {
	cfg EmailConfig
}

func NewEmailService(cfg EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendPasswordReset(toEmail, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.AppURL, resetToken)

	subject := "Reset your Trackr password"
	body := fmt.Sprintf(`Hi,

You requested a password reset for your Trackr account (%s).

Click the link below to set a new password. This link expires in 1 hour.

%s

If you did not request this, you can safely ignore this email.
Your password will not be changed.

— The Trackr Team
`, toEmail, resetURL)

	return s.send(toEmail, subject, body)
}

func (s *EmailService) SendReminderAlert(toEmail, role, company, status string, silentDays int) error {
	subject := fmt.Sprintf("Follow-up reminder: %s at %s", role, company)
	body := fmt.Sprintf(`Hi,

It has been %d day(s) since you applied for the %s position at %s with no update (current status: %s).

Now might be a good time to send a follow-up email to check on your application.

View your application: %s

Good luck!

— The Trackr Team
`, silentDays, role, company, status, s.cfg.AppURL)

	return s.send(toEmail, subject, body)
}

func (s *EmailService) SendWelcome(toEmail string) error {
	subject := "Welcome to Trackr"
	body := fmt.Sprintf(`Hi,

Your Trackr account has been created for %s.

Start tracking your job applications at: %s

Good luck with your search!

— The Trackr Team
`, toEmail, s.cfg.AppURL)

	return s.send(toEmail, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", s.cfg.FromName, s.cfg.FromAddress),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, s.cfg.FromAddress, []string{to}, []byte(msg))
}

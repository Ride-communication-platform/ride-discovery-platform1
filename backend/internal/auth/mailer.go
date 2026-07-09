package auth

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/resend/resend-go/v3"
)

type Mailer interface {
	SendVerificationCode(toEmail, toName, code string) error
	SendPasswordResetCode(toEmail, toName, code string) error
}

type SMTPMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPMailer(host, port, user, pass, from string) Mailer {
	if host == "" || port == "" || user == "" || pass == "" || from == "" {
		return nil
	}

	return &SMTPMailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (m *SMTPMailer) SendVerificationCode(toEmail, toName, code string) error {
	subject := "GenRide email verification"
	body := fmt.Sprintf(
		"Hi %s,\n\nYour GenRide verification code is: %s\n\nEnter this code in the app to verify your email.\n",
		safeDisplayName(toName),
		code,
	)
	return m.send(toEmail, subject, body)
}

func (m *SMTPMailer) SendPasswordResetCode(toEmail, toName, code string) error {
	subject := "GenRide password reset"
	body := fmt.Sprintf(
		"Hi %s,\n\nYour GenRide password reset code is: %s\n\nEnter this code in the app to reset your password. If you did not request this, ignore this email.\n",
		safeDisplayName(toName),
		code,
	)
	return m.send(toEmail, subject, body)
}

func (m *SMTPMailer) send(toEmail, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", m.from),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	return smtp.SendMail(addr, auth, m.from, []string{toEmail}, []byte(message))
}

type ResendMailer struct {
	client *resend.Client
	from   string
}

func NewResendMailer(apiKey, from string) Mailer {
	apiKey = strings.TrimSpace(apiKey)
	from = strings.TrimSpace(from)

	if apiKey == "" || from == "" {
		return nil
	}

	return &ResendMailer{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

func (m *ResendMailer) SendVerificationCode(toEmail, toName, code string) error {
	subject := "GenRide email verification"

	html := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your GenRide verification code is:</p>
		<h2 style="letter-spacing: 2px;">%s</h2>
		<p>Enter this code in the app to verify your email.</p>
	`, escapeHTML(safeDisplayName(toName)), escapeHTML(code))

	return m.send(toEmail, subject, html)
}

func (m *ResendMailer) SendPasswordResetCode(toEmail, toName, code string) error {
	subject := "GenRide password reset"

	html := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your GenRide password reset code is:</p>
		<h2 style="letter-spacing: 2px;">%s</h2>
		<p>Enter this code in the app to reset your password.</p>
		<p>If you did not request this, ignore this email.</p>
	`, escapeHTML(safeDisplayName(toName)), escapeHTML(code))

	return m.send(toEmail, subject, html)
}

func (m *ResendMailer) send(toEmail, subject, html string) error {
	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{toEmail},
		Subject: subject,
		Html:    html,
	}

	_, err := m.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send resend email: %w", err)
	}

	return nil
}

func safeDisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "there"
	}
	return trimmed
}

func escapeHTML(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	value = strings.ReplaceAll(value, `"`, "&quot;")
	value = strings.ReplaceAll(value, "'", "&#39;")
	return value
}

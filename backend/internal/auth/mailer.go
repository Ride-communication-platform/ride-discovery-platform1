package auth

import (
	"fmt"
	"net/smtp"
	"strings"
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
	subject := "RideX email verification"
	body := fmt.Sprintf("Hi %s,\n\nYour RideX verification code is: %s\n\nEnter this code in the app to verify your email.\n", safeDisplayName(toName), code)
	return m.send(toEmail, subject, body)
}

func (m *SMTPMailer) SendPasswordResetCode(toEmail, toName, code string) error {
	subject := "RideX password reset"
	body := fmt.Sprintf("Hi %s,\n\nYour RideX password reset code is: %s\n\nEnter this code in the app to reset your password. If you did not request this, ignore this email.\n", safeDisplayName(toName), code)
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

func safeDisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "there"
	}
	return trimmed
}

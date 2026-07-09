package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"net/smtp"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

type GmailAPIMailer struct {
	from       string
	httpClient *http.Client
}

func NewGmailAPIMailer(clientID, clientSecret, refreshToken, from string) Mailer {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	refreshToken = strings.TrimSpace(refreshToken)
	from = strings.TrimSpace(from)

	if clientID == "" || clientSecret == "" || refreshToken == "" || from == "" {
		return nil
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.send"},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	return &GmailAPIMailer{
		from:       from,
		httpClient: config.Client(context.Background(), token),
	}
}

func (m *GmailAPIMailer) SendVerificationCode(toEmail, toName, code string) error {
	subject := "GenRide email verification"

	htmlBody := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your GenRide verification code is:</p>
		<h2 style="letter-spacing:2px;">%s</h2>
		<p>Enter this code in the app to verify your email.</p>
	`, html.EscapeString(safeDisplayName(toName)), html.EscapeString(code))

	return m.send(toEmail, subject, htmlBody)
}

func (m *GmailAPIMailer) SendPasswordResetCode(toEmail, toName, code string) error {
	subject := "GenRide password reset"

	htmlBody := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your GenRide password reset code is:</p>
		<h2 style="letter-spacing:2px;">%s</h2>
		<p>Enter this code in the app to reset your password.</p>
		<p>If you did not request this, ignore this email.</p>
	`, html.EscapeString(safeDisplayName(toName)), html.EscapeString(code))

	return m.send(toEmail, subject, htmlBody)
}

func (m *GmailAPIMailer) send(toEmail, subject, htmlBody string) error {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf("From: GenRide <%s>\r\n", m.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.TrimSpace(toEmail)))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	raw := base64.RawURLEncoding.EncodeToString(msg.Bytes())

	requestBody := fmt.Sprintf(`{"raw":"%s"}`, raw)

	req, err := http.NewRequest(
		http.MethodPost,
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
		strings.NewReader(requestBody),
	)
	if err != nil {
		return fmt.Errorf("create gmail api request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send gmail api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gmail api send failed with status %s", resp.Status)
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

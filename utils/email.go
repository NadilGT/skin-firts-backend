package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"
)

// SendEmail sends an HTML email using SMTP with explicit STARTTLS on port 587.
// All errors are fully logged so they appear in Render's log stream.
func SendEmail(to []string, subject string, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	// ── Validate env vars and log presence (never log the full password) ──────
	log.Printf("[EMAIL] SMTP_HOST=%q  SMTP_PORT=%q  SMTP_USER=%q  SMTP_PASS length=%d",
		smtpHost, smtpPort, smtpUser, len(smtpPass))

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("[EMAIL] one or more SMTP env vars are empty — host:%q port:%q user:%q pass_len:%d",
			smtpHost, smtpPort, smtpUser, len(smtpPass))
	}

	addr := smtpHost + ":" + smtpPort
	log.Printf("[EMAIL] Dialing %s (plain TCP, will upgrade to STARTTLS)…", addr)

	// ── Step 1: dial with a 15-second timeout ────────────────────────────────
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("[EMAIL] TCP dial to %s failed: %w  — Render may be blocking outbound port %s", addr, err, smtpPort)
	}
	defer conn.Close()
	log.Printf("[EMAIL] TCP connection to %s established", addr)

	// ── Step 2: create SMTP client ───────────────────────────────────────────
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("[EMAIL] smtp.NewClient failed: %w", err)
	}
	defer client.Close()

	// ── Step 3: upgrade to TLS via STARTTLS ──────────────────────────────────
	tlsCfg := &tls.Config{
		ServerName: smtpHost,
		MinVersion: tls.VersionTLS12,
	}
	log.Printf("[EMAIL] Sending STARTTLS (ServerName=%s)…", smtpHost)
	if err = client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("[EMAIL] STARTTLS negotiation failed: %w", err)
	}
	log.Printf("[EMAIL] STARTTLS OK — TLS tunnel established")

	// ── Step 4: authenticate ─────────────────────────────────────────────────
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	log.Printf("[EMAIL] Authenticating as %s…", smtpUser)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("[EMAIL] SMTP AUTH failed (check Gmail App Password — no spaces): %w", err)
	}
	log.Printf("[EMAIL] SMTP authentication successful")

	// ── Step 5: set sender and recipient ────────────────────────────────────
	if err = client.Mail(smtpUser); err != nil {
		return fmt.Errorf("[EMAIL] MAIL FROM failed: %w", err)
	}
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("[EMAIL] RCPT TO <%s> failed: %w", recipient, err)
		}
	}

	// ── Step 6: write the message body ──────────────────────────────────────
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("[EMAIL] DATA command failed: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		smtpUser, to[0], subject, body,
	)
	if _, err = fmt.Fprint(wc, msg); err != nil {
		return fmt.Errorf("[EMAIL] writing message body failed: %w", err)
	}
	if err = wc.Close(); err != nil {
		return fmt.Errorf("[EMAIL] closing DATA writer failed: %w", err)
	}

	// ── Step 7: QUIT ─────────────────────────────────────────────────────────
	if err = client.Quit(); err != nil {
		// Non-fatal — message was already accepted by the server
		log.Printf("[EMAIL] QUIT warning (non-fatal): %v", err)
	}

	log.Printf("[EMAIL] ✅ Email delivered to %v via %s", to, addr)
	return nil
}

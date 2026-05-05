package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendEmail sends an HTML email via the Resend API (HTTPS port 443).
// This bypasses all SMTP port restrictions on cloud platforms like Render.
//
// Required env var: RESEND_API_KEY
// Optional env var: EMAIL_FROM  (defaults to onboarding@resend.dev for testing)
func SendEmail(to []string, subject string, body string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("[EMAIL] RESEND_API_KEY env var is not set")
	}

	// Use a custom verified sender domain if configured, otherwise fall back
	// to Resend's shared testing address (works without a domain, to: must be
	// your own address while in testing mode).
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "onboarding@resend.dev"
	}

	log.Printf("[EMAIL] Sending via Resend → from=%s  to=%v  subject=%q", from, to, subject)

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      to,
		Subject: subject,
		Html:    body,
	}

	resp, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("[EMAIL] Resend API error: %w", err)
	}

	log.Printf("[EMAIL] ✅ Email sent via Resend — message ID: %s", resp.Id)
	return nil
}

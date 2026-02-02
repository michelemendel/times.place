package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const resendAPIURL = "https://api.resend.com/emails"

// ResendSender sends verification emails via the Resend API.
type ResendSender struct {
	apiKey string
	from   string
	client *http.Client
}

// NewResendSender creates a Resend sender using RESEND_API_KEY (and optional RESEND_FROM) from env.
func NewResendSender() *ResendSender {
	from := os.Getenv("RESEND_FROM")
	if from == "" {
		from = "Times.Place <onboarding@resend.dev>"
	}
	return &ResendSender{
		apiKey: os.Getenv("RESEND_API_KEY"),
		from:   from,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// sendEmailRequest matches Resend API request body.
type sendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendVerificationEmail sends a verification email with the given link.
func (r *ResendSender) SendVerificationEmail(to, verificationLink string) error {
	if r.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}
	subject := "Verify your email – Times.Place"
	html := fmt.Sprintf(`<p>Please verify your email by clicking the link below:</p>
<p><a href="%s">%s</a></p>
<p>This link expires in 24 hours. If you did not sign up for Times.Place, you can ignore this email.</p>`,
		verificationLink, verificationLink)
	body := sendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errData map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return fmt.Errorf("resend API returned %d: %v", resp.StatusCode, errData)
	}
	return nil
}

// SendPasswordResetEmail sends a password reset email with the given link.
func (r *ResendSender) SendPasswordResetEmail(to, resetLink string) error {
	if r.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}
	subject := "Reset your password – Times.Place"
	html := fmt.Sprintf(`<p>You requested to reset your password. Please click the link below:</p>
<p><a href="%s">%s</a></p>
<p>This link expires in 1 hour. If you did not request a password reset, you can safely ignore this email.</p>`,
		resetLink, resetLink)
	body := sendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errData map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return fmt.Errorf("resend API returned %d: %v", resp.StatusCode, errData)
	}
	return nil
}

package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"messaging-service/internal/domain"
)

// SendGridEmailProvider implements domain.EmailProvider for SendGrid
type SendGridEmailProvider struct {
	apiKey     string
	httpClient *http.Client
	shouldFail bool
	errorCode  int
}

// NewSendGridEmailProvider creates a new SendGrid email provider
func NewSendGridEmailProvider(apiKey string) *SendGridEmailProvider {
	return &SendGridEmailProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendEmail sends an email through SendGrid
func (p *SendGridEmailProvider) SendEmail(ctx context.Context, from, to, body string, attachments []string) error {
	// Simulate provider errors for testing
	if p.shouldFail {
		return &domain.ProviderError{
			Code:    p.errorCode,
			Message: fmt.Sprintf("SendGrid error: %d", p.errorCode),
		}
	}

	// In a real implementation, you would:
	// 1. Create the SendGrid API request
	// 2. Add attachments if provided
	// 3. Send the request to SendGrid API
	// 4. Handle the response

	// For now, we'll just simulate success
	fmt.Printf("SendGrid: Sending email from %s to %s\n", from, to)
	return nil
}

// SetFailureMode sets the provider to fail with specific error code (for testing)
func (p *SendGridEmailProvider) SetFailureMode(shouldFail bool, errorCode int) {
	p.shouldFail = shouldFail
	p.errorCode = errorCode
}

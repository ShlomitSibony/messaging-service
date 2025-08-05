package provider

import (
	"messaging-service/internal/domain"
)

// EmailProviderType represents the type of email provider
type EmailProviderType string

const (
	EmailProviderMock     EmailProviderType = "mock"
	EmailProviderSendGrid EmailProviderType = "sendgrid"
)

// NewEmailProvider creates an email provider based on the specified type
func NewEmailProvider(providerType EmailProviderType, config map[string]string) domain.EmailProvider {
	switch providerType {
	case EmailProviderSendGrid:
		apiKey := config["api_key"]
		return NewSendGridEmailProvider(apiKey)
	case EmailProviderMock:
		fallthrough
	default:
		return NewMockEmailProvider()
	}
}

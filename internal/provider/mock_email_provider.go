package provider

import (
	"context"
	"fmt"
	"messaging-service/internal/domain"
	"sync"
	"time"
)

type MockEmailProvider struct {
	messages   []MockEmailMessage
	mu         sync.RWMutex
	shouldFail bool
	errorCode  int
}

type MockEmailMessage struct {
	From        string
	To          string
	Body        string
	Attachments []string
	Timestamp   time.Time
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider() domain.EmailProvider {
	return &MockEmailProvider{
		messages: make([]MockEmailMessage, 0),
	}
}

// NewMockEmailProviderWithFailure creates a mock email provider that fails
func NewMockEmailProviderWithFailure() domain.EmailProvider {
	return &MockEmailProvider{
		messages:   make([]MockEmailMessage, 0),
		shouldFail: true,
		errorCode:  500,
	}
}

// NewMockEmailProviderWithErrorCode creates a mock email provider that fails with specific HTTP error code
func NewMockEmailProviderWithErrorCode(errorCode int) domain.EmailProvider {
	return &MockEmailProvider{
		messages:   make([]MockEmailMessage, 0),
		shouldFail: true,
		errorCode:  errorCode,
	}
}

func (p *MockEmailProvider) SendEmail(ctx context.Context, from, to, body string, attachments []string) error {
	// Handle specific error codes
	if p.shouldFail {
		switch p.errorCode {
		case 500:
			return &domain.ProviderError{
				Code:    500,
				Message: "Internal server error",
			}
		case 429:
			return &domain.ProviderError{
				Code:       429,
				Message:    "Too many requests",
				RetryAfter: 30, // 30 seconds
			}
		default:
			return fmt.Errorf("mock email provider failure")
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	message := MockEmailMessage{
		From:        from,
		To:          to,
		Body:        body,
		Attachments: attachments,
		Timestamp:   time.Now(),
	}

	p.messages = append(p.messages, message)
	return nil
}

// GetMessages returns all sent messages (for testing)
func (p *MockEmailProvider) GetMessages() []MockEmailMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	messages := make([]MockEmailMessage, len(p.messages))
	copy(messages, p.messages)
	return messages
}

// ClearMessages clears all sent messages (for testing)
func (p *MockEmailProvider) ClearMessages() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make([]MockEmailMessage, 0)
}

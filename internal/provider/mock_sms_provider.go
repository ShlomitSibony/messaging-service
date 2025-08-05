package provider

import (
	"context"
	"fmt"
	"messaging-service/internal/domain"
	"sync"
	"time"
)

type MockSMSProvider struct {
	messages   []MockSMSMessage
	mu         sync.RWMutex
	shouldFail bool
	errorCode  int
}

type MockSMSMessage struct {
	From        string
	To          string
	Body        string
	Attachments []string
	Timestamp   time.Time
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider() domain.SMSProvider {
	return &MockSMSProvider{
		messages: make([]MockSMSMessage, 0),
	}
}

// NewMockSMSProviderWithFailure creates a mock SMS provider that fails
func NewMockSMSProviderWithFailure() domain.SMSProvider {
	return &MockSMSProvider{
		messages:   make([]MockSMSMessage, 0),
		shouldFail: true,
		errorCode:  500,
	}
}

// NewMockSMSProviderWithErrorCode creates a mock SMS provider that fails with specific HTTP error code
func NewMockSMSProviderWithErrorCode(errorCode int) domain.SMSProvider {
	return &MockSMSProvider{
		messages:   make([]MockSMSMessage, 0),
		shouldFail: true,
		errorCode:  errorCode,
	}
}

func (p *MockSMSProvider) SendSMS(ctx context.Context, from, to, body string) error {
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
			return fmt.Errorf("mock SMS provider failure")
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	message := MockSMSMessage{
		From:      from,
		To:        to,
		Body:      body,
		Timestamp: time.Now(),
	}

	p.messages = append(p.messages, message)
	return nil
}

func (p *MockSMSProvider) SendMMS(ctx context.Context, from, to, body string, attachments []string) error {
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
			return fmt.Errorf("mock MMS provider failure")
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	message := MockSMSMessage{
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
func (p *MockSMSProvider) GetMessages() []MockSMSMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	messages := make([]MockSMSMessage, len(p.messages))
	copy(messages, p.messages)
	return messages
}

// ClearMessages clears all sent messages (for testing)
func (p *MockSMSProvider) ClearMessages() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make([]MockSMSMessage, 0)
}

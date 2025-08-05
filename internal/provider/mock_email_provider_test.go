package provider

import (
	"context"
	"testing"

	"messaging-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestMockEmailProvider_SendEmail(t *testing.T) {
	provider := NewMockEmailProvider()

	ctx := context.Background()
	from := "user@usehatchapp.com"
	to := "contact@gmail.com"
	body := "Test email message"
	attachments := []string{"https://example.com/document.pdf"}

	err := provider.SendEmail(ctx, from, to, body, attachments)
	assert.NoError(t, err)

	mockProvider := provider.(*MockEmailProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, from, messages[0].From)
	assert.Equal(t, to, messages[0].To)
	assert.Equal(t, body, messages[0].Body)
	assert.Equal(t, attachments, messages[0].Attachments)
}

func TestMockEmailProvider_WithFailure(t *testing.T) {
	provider := NewMockEmailProviderWithFailure()

	ctx := context.Background()
	from := "user@usehatchapp.com"
	to := "contact@gmail.com"
	body := "Test email message"
	attachments := []string{}

	err := provider.SendEmail(ctx, from, to, body, attachments)
	assert.Error(t, err)

	// Should return a ProviderError with 500 status
	if providerErr, ok := err.(*domain.ProviderError); ok {
		assert.Equal(t, 500, providerErr.Code)
		assert.Equal(t, "Internal server error", providerErr.Message)
	} else {
		t.Fatal("Expected ProviderError but got different error type")
	}

	mockProvider := provider.(*MockEmailProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 0)
}

func TestMockEmailProvider_ClearMessages(t *testing.T) {
	provider := NewMockEmailProvider()

	ctx := context.Background()

	// Send a message
	err := provider.SendEmail(ctx, "user@usehatchapp.com", "contact@gmail.com", "Test message", []string{})
	assert.NoError(t, err)

	// Verify message was sent
	mockProvider := provider.(*MockEmailProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 1)

	// Clear messages
	mockProvider.ClearMessages()

	// Verify messages were cleared
	messages = mockProvider.GetMessages()
	assert.Len(t, messages, 0)
}

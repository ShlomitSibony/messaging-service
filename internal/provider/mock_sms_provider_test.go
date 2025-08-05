package provider

import (
	"context"
	"testing"

	"messaging-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestMockSMSProvider_SendSMS_Success(t *testing.T) {
	provider := NewMockSMSProvider()
	ctx := context.Background()

	err := provider.SendSMS(ctx, "+1234567890", "+0987654321", "Hello, World!")
	assert.NoError(t, err)

	mockProvider := provider.(*MockSMSProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "+1234567890", messages[0].From)
	assert.Equal(t, "+0987654321", messages[0].To)
	assert.Equal(t, "Hello, World!", messages[0].Body)
}

func TestMockSMSProvider_SendSMS_WithFailure(t *testing.T) {
	provider := NewMockSMSProviderWithFailure()
	ctx := context.Background()

	err := provider.SendSMS(ctx, "+1234567890", "+0987654321", "Hello, World!")
	assert.Error(t, err)

	// Should return a ProviderError with 500 status
	if providerErr, ok := err.(*domain.ProviderError); ok {
		assert.Equal(t, 500, providerErr.Code)
		assert.Equal(t, "Internal server error", providerErr.Message)
	} else {
		t.Fatal("Expected ProviderError but got different error type")
	}
}

func TestMockSMSProvider_SendSMS_WithSpecificErrorCode(t *testing.T) {
	provider := NewMockSMSProviderWithErrorCode(429)
	ctx := context.Background()

	err := provider.SendSMS(ctx, "+1234567890", "+0987654321", "Hello, World!")
	assert.Error(t, err)

	// Should return a ProviderError with 429 status
	if providerErr, ok := err.(*domain.ProviderError); ok {
		assert.Equal(t, 429, providerErr.Code)
		assert.Equal(t, "Too many requests", providerErr.Message)
		assert.Equal(t, 30, providerErr.RetryAfter)
	} else {
		t.Fatal("Expected ProviderError but got different error type")
	}
}

func TestMockSMSProvider_SendMMS_Success(t *testing.T) {
	provider := NewMockSMSProvider()
	ctx := context.Background()
	attachments := []string{"image1.jpg", "image2.png"}

	err := provider.SendMMS(ctx, "+1234567890", "+0987654321", "Hello, World!", attachments)
	assert.NoError(t, err)

	mockProvider := provider.(*MockSMSProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "+1234567890", messages[0].From)
	assert.Equal(t, "+0987654321", messages[0].To)
	assert.Equal(t, "Hello, World!", messages[0].Body)
	assert.Equal(t, attachments, messages[0].Attachments)
}

func TestMockSMSProvider_SendMMS_WithFailure(t *testing.T) {
	provider := NewMockSMSProviderWithFailure()
	ctx := context.Background()
	attachments := []string{"image1.jpg"}

	err := provider.SendMMS(ctx, "+1234567890", "+0987654321", "Hello, World!", attachments)
	assert.Error(t, err)

	if providerErr, ok := err.(*domain.ProviderError); ok {
		assert.Equal(t, 500, providerErr.Code)
		assert.Equal(t, "Internal server error", providerErr.Message)
	} else {
		t.Fatal("Expected ProviderError but got different error type")
	}
}

func TestMockSMSProvider_ClearMessages(t *testing.T) {
	provider := NewMockSMSProvider()
	ctx := context.Background()

	// Send a message
	err := provider.SendSMS(ctx, "+1234567890", "+0987654321", "Hello, World!")
	assert.NoError(t, err)

	// Verify message was sent
	mockProvider := provider.(*MockSMSProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 1)

	// Clear messages
	mockProvider.ClearMessages()

	// Verify messages were cleared
	messages = mockProvider.GetMessages()
	assert.Len(t, messages, 0)
}

func TestDomain_IsRetryableError(t *testing.T) {
	// Test retryable errors
	retryableCodes := []int{429, 500, 502, 503, 504}
	for _, code := range retryableCodes {
		err := &domain.ProviderError{Code: code, Message: "test"}
		assert.True(t, domain.IsRetryableError(err), "Error code %d should be retryable", code)
	}

	// Test non-retryable errors
	nonRetryableCodes := []int{400, 401, 403, 404, 422}
	for _, code := range nonRetryableCodes {
		err := &domain.ProviderError{Code: code, Message: "test"}
		assert.False(t, domain.IsRetryableError(err), "Error code %d should not be retryable", code)
	}

	// Test non-ProviderError
	regularErr := assert.AnError
	assert.False(t, domain.IsRetryableError(regularErr))
}

func TestDomain_GetRetryAfterSeconds(t *testing.T) {
	// Test rate limit error with RetryAfter
	err := &domain.ProviderError{Code: 429, Message: "rate limited", RetryAfter: 30}
	assert.Equal(t, 30, domain.GetRetryAfterSeconds(err))

	// Test non-rate limit error
	err = &domain.ProviderError{Code: 500, Message: "server error"}
	assert.Equal(t, 0, domain.GetRetryAfterSeconds(err))

	// Test non-ProviderError
	regularErr := assert.AnError
	assert.Equal(t, 0, domain.GetRetryAfterSeconds(regularErr))
}

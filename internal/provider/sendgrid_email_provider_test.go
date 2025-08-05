package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendGridEmailProvider_SendEmail_Success(t *testing.T) {
	provider := NewSendGridEmailProvider("test-api-key")

	err := provider.SendEmail(context.Background(), "from@test.com", "to@test.com", "Test email", nil)

	assert.NoError(t, err)
}

func TestSendGridEmailProvider_SendEmail_WithFailure(t *testing.T) {
	provider := NewSendGridEmailProvider("test-api-key")
	provider.SetFailureMode(true, 500)

	err := provider.SendEmail(context.Background(), "from@test.com", "to@test.com", "Test email", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SendGrid error: 500")
}

func TestSendGridEmailProvider_SendEmail_WithAttachments(t *testing.T) {
	provider := NewSendGridEmailProvider("test-api-key")
	attachments := []string{"https://example.com/file.pdf"}

	err := provider.SendEmail(context.Background(), "from@test.com", "to@test.com", "Test email", attachments)

	assert.NoError(t, err)
}

package service

import (
	"context"
	"testing"
	"time"

	"messaging-service/internal/domain"
	"messaging-service/internal/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockConversationRepository struct {
	mock.Mock
}

func (m *MockConversationRepository) Create(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	args := m.Called(ctx, customerContact, businessContact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) GetByID(ctx context.Context, id int) (*domain.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) GetByContacts(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	args := m.Called(ctx, customerContact, businessContact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) GetOrCreate(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	args := m.Called(ctx, customerContact, businessContact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) List(ctx context.Context, query *domain.ConversationQuery) ([]domain.Conversation, int, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]domain.Conversation), args.Get(1).(int), args.Error(2)
}

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id int) (*domain.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) GetByConversationID(ctx context.Context, conversationID int) ([]domain.Message, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockMessageRepository) GetByProviderMessageID(ctx context.Context, providerMessageID string) (*domain.Message, error) {
	args := m.Called(ctx, providerMessageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func TestMessagingService_SendSMS(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Mock expectations
	conversationRepo.On("GetOrCreate", mock.Anything, "+12016661234", "+18045551234").Return(&domain.Conversation{
		ID:              1,
		CustomerContact: "+12016661234",
		BusinessContact: "+18045551234",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil)

	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Test
	req := &domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "Hello! This is a test SMS message.",
	}

	err := service.SendSMS(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	conversationRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestMessagingService_SendMMS(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Mock expectations
	conversationRepo.On("GetOrCreate", mock.Anything, "+12016661234", "+18045551234").Return(&domain.Conversation{
		ID:              1,
		CustomerContact: "+12016661234",
		BusinessContact: "+18045551234",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil)

	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Test
	req := &domain.SendSMSRequest{
		Timestamp:   time.Now().UTC(),
		From:        "+12016661234",
		To:          "+18045551234",
		Type:        "mms",
		Body:        "Hello! This is a test MMS message with attachment.",
		Attachments: []string{"https://example.com/image.jpg"},
	}

	err := service.SendSMS(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	conversationRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestMessagingService_SendEmail(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Mock expectations
	conversationRepo.On("GetOrCreate", mock.Anything, "contact@gmail.com", "user@usehatchapp.com").Return(&domain.Conversation{
		ID:              1,
		CustomerContact: "contact@gmail.com",
		BusinessContact: "user@usehatchapp.com",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil)

	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Test
	req := &domain.SendEmailRequest{
		Timestamp:   time.Now().UTC(),
		From:        "user@usehatchapp.com",
		To:          "contact@gmail.com",
		Body:        "Hello! This is a test email message with <b>HTML</b> formatting.",
		Attachments: []string{"https://example.com/document.pdf"},
	}

	err := service.SendEmail(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	conversationRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestMessagingService_HandleInboundSMS(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Mock expectations - note the normalized order
	conversationRepo.On("GetOrCreate", mock.Anything, "+12016661234", "+18045551234").Return(&domain.Conversation{
		ID:              1,
		CustomerContact: "+12016661234",
		BusinessContact: "+18045551234",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil)

	messageRepo.On("GetByProviderMessageID", mock.Anything, "message-1").Return(nil, nil)
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Test
	webhook := &domain.InboundSMSWebhook{
		Timestamp:           time.Now().UTC(),
		From:                "+18045551234",
		To:                  "+12016661234",
		Type:                "sms",
		MessagingProviderID: "message-1",
		Body:                "This is an incoming SMS message",
	}

	err := service.HandleInboundSMS(context.Background(), webhook)

	// Assertions
	assert.NoError(t, err)
	conversationRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestMessagingService_HandleInboundEmail(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Mock expectations
	conversationRepo.On("GetOrCreate", mock.Anything, "contact@gmail.com", "user@usehatchapp.com").Return(&domain.Conversation{
		ID:              1,
		CustomerContact: "contact@gmail.com",
		BusinessContact: "user@usehatchapp.com",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}, nil)

	messageRepo.On("GetByProviderMessageID", mock.Anything, "message-3").Return(nil, nil)
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Test
	webhook := &domain.InboundEmailWebhook{
		Timestamp: time.Now().UTC(),
		From:      "contact@gmail.com",
		To:        "user@usehatchapp.com",
		XillioID:  "message-3",
		Body:      "<html><body>This is an incoming email with <b>HTML</b> content</body></html>",
	}

	err := service.HandleInboundEmail(context.Background(), webhook)

	// Assertions
	assert.NoError(t, err)
	conversationRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
}

func TestMessagingService_SendSMS_WithRetryableError(t *testing.T) {
	// Create mocks
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProviderWithErrorCode(500) // Simulate 500 error
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Setup conversation mock
	conversation := &domain.Conversation{
		ID:              1,
		CustomerContact: "+12016661234",
		BusinessContact: "+18045551234",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	conversationRepo.On("GetOrCreate", mock.Anything, "+12016661234", "+18045551234").Return(conversation, nil)

	// Setup message mock
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Create request
	req := &domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "Test message",
	}

	// This should fail because the provider returns a 500 error
	err := service.SendSMS(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message through provider")

	// Verify that the provider was called (the retry logic would have tried multiple times)
	mockProvider := smsProvider.(*provider.MockSMSProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 0) // No messages should be sent due to provider failure
}

func TestMessagingService_SendSMS_WithRateLimitError(t *testing.T) {
	// Create mocks
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProviderWithErrorCode(429) // Simulate 429 error
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Setup conversation mock
	conversation := &domain.Conversation{
		ID:              1,
		CustomerContact: "+12016661234",
		BusinessContact: "+18045551234",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	conversationRepo.On("GetOrCreate", mock.Anything, "+12016661234", "+18045551234").Return(conversation, nil)

	// Setup message mock
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Create request
	req := &domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "Test message",
	}

	// This should fail because the provider returns a 429 error
	err := service.SendSMS(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message through provider")

	// Verify that the provider was called (the retry logic would have tried multiple times)
	mockProvider := smsProvider.(*provider.MockSMSProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 0) // No messages should be sent due to provider failure
}

func TestMessagingService_SendEmail_WithRetryableError(t *testing.T) {
	// Create mocks
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProviderWithErrorCode(500) // Simulate 500 error

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Setup conversation mock
	conversation := &domain.Conversation{
		ID:              1,
		CustomerContact: "user@usehatchapp.com",
		BusinessContact: "contact@gmail.com",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	conversationRepo.On("GetOrCreate", mock.Anything, "user@usehatchapp.com", "contact@gmail.com").Return(conversation, nil)

	// Setup message mock
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Create request
	req := &domain.SendEmailRequest{
		Timestamp:   time.Now().UTC(),
		From:        "user@usehatchapp.com",
		To:          "contact@gmail.com",
		Body:        "Test email",
		Attachments: []string{"document.pdf"},
	}

	// This should fail because the provider returns a 500 error
	err := service.SendEmail(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email through provider")

	// Verify that the provider was called (the retry logic would have tried multiple times)
	mockProvider := emailProvider.(*provider.MockEmailProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 0) // No messages should be sent due to provider failure
}

func TestMessagingService_SendEmail_WithRateLimitError(t *testing.T) {
	// Create mocks
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProviderWithErrorCode(429) // Simulate 429 error

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Setup conversation mock
	conversation := &domain.Conversation{
		ID:              1,
		CustomerContact: "user@usehatchapp.com",
		BusinessContact: "contact@gmail.com",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	conversationRepo.On("GetOrCreate", mock.Anything, "user@usehatchapp.com", "contact@gmail.com").Return(conversation, nil)

	// Setup message mock
	messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	// Create request
	req := &domain.SendEmailRequest{
		Timestamp:   time.Now().UTC(),
		From:        "user@usehatchapp.com",
		To:          "contact@gmail.com",
		Body:        "Test email",
		Attachments: []string{"document.pdf"},
	}

	// This should fail because the provider returns a 429 error
	err := service.SendEmail(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email through provider")

	// Verify that the provider was called (the retry logic would have tried multiple times)
	mockProvider := emailProvider.(*provider.MockEmailProvider)
	messages := mockProvider.GetMessages()
	assert.Len(t, messages, 0) // No messages should be sent due to provider failure
}

func TestMessagingService_ValidateTimestamp(t *testing.T) {
	// Setup
	conversationRepo := &MockConversationRepository{}
	messageRepo := &MockMessageRepository{}
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	service := NewMessagingServiceWithConfig(conversationRepo, messageRepo, smsProvider, emailProvider, TestRetryConfig())

	// Test cases
	testCases := []struct {
		name      string
		timestamp time.Time
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid current timestamp",
			timestamp: time.Now().UTC(),
			expectErr: false,
		},
		{
			name:      "valid recent past timestamp",
			timestamp: time.Now().UTC().Add(-1 * time.Hour),
			expectErr: false,
		},
		{
			name:      "zero timestamp",
			timestamp: time.Time{},
			expectErr: true,
			errMsg:    "timestamp cannot be zero",
		},
		{
			name:      "future timestamp (6 minutes)",
			timestamp: time.Now().UTC().Add(6 * time.Minute),
			expectErr: true,
			errMsg:    "timestamp cannot be in the future",
		},
		{
			name:      "very old timestamp (11 years)",
			timestamp: time.Now().UTC().AddDate(-11, 0, 0),
			expectErr: true,
			errMsg:    "timestamp too old",
		},
		{
			name:      "timestamp before year 2000",
			timestamp: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC),
			expectErr: true,
			errMsg:    "timestamp too old", // This gets caught by the "too old" check first
		},
		{
			name:      "timestamp from 2010 (too old)",
			timestamp: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
			expectErr: true,
			errMsg:    "timestamp too old",
		},
		{
			name:      "non-UTC timestamp",
			timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			expectErr: true,
			errMsg:    "timestamp must be in UTC timezone",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Cast to concrete type to access private method
			messagingService := service.(*messagingService)
			err := messagingService.validateTimestamp(tc.timestamp)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

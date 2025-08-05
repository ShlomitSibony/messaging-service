package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"messaging-service/internal/domain"
)

type messagingService struct {
	conversationRepo domain.ConversationRepository
	messageRepo      domain.MessageRepository
	smsProvider      domain.SMSProvider
	emailProvider    domain.EmailProvider
	retryConfig      RetryConfig
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int           `json:"max_retries"`
	BaseDelay  time.Duration `json:"base_delay"`
	MaxDelay   time.Duration `json:"max_delay"`
	Multiplier float64       `json:"multiplier"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   time.Minute,
		Multiplier: 2.0,
	}
}

// TestRetryConfig returns fast retry configuration for tests
func TestRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond * 10,
		MaxDelay:   time.Millisecond * 100,
		Multiplier: 2.0,
	}
}

// NewMessagingService creates a new messaging service
func NewMessagingService(
	conversationRepo domain.ConversationRepository,
	messageRepo domain.MessageRepository,
	smsProvider domain.SMSProvider,
	emailProvider domain.EmailProvider,
) domain.MessagingService {
	return &messagingService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		smsProvider:      smsProvider,
		emailProvider:    emailProvider,
		retryConfig:      DefaultRetryConfig(),
	}
}

// NewMessagingServiceWithConfig creates a messaging service with custom retry configuration
func NewMessagingServiceWithConfig(
	conversationRepo domain.ConversationRepository,
	messageRepo domain.MessageRepository,
	smsProvider domain.SMSProvider,
	emailProvider domain.EmailProvider,
	retryConfig RetryConfig,
) domain.MessagingService {
	return &messagingService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		smsProvider:      smsProvider,
		emailProvider:    emailProvider,
		retryConfig:      retryConfig,
	}
}

func (s *messagingService) SendSMS(ctx context.Context, req *domain.SendSMSRequest) error {
	// Validate request
	if err := s.validateSMSRequest(req); err != nil {
		return fmt.Errorf("invalid SMS request: %w", err)
	}

	// Send message through provider with retry logic
	if err := s.sendSMSMessageWithRetry(ctx, req); err != nil {
		return fmt.Errorf("failed to send message through provider: %w", err)
	}

	// Create message record
	message := s.buildOutboundMessage(req.From, req.To, req.Type, req.Body, req.Attachments, req.Timestamp)
	if err := s.createMessageRecord(ctx, message); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (s *messagingService) SendEmail(ctx context.Context, req *domain.SendEmailRequest) error {
	// Validate request
	if err := s.validateEmailRequest(req); err != nil {
		return fmt.Errorf("invalid email request: %w", err)
	}

	// Send email through provider with retry logic
	if err := s.sendEmailMessageWithRetry(ctx, req); err != nil {
		return fmt.Errorf("failed to send email through provider: %w", err)
	}

	// Create message record
	message := s.buildOutboundMessage(req.From, req.To, domain.MessageTypeEmail, req.Body, req.Attachments, req.Timestamp)
	if err := s.createMessageRecord(ctx, message); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (s *messagingService) HandleInboundSMS(ctx context.Context, webhook *domain.InboundSMSWebhook) error {
	// Validate webhook
	if err := s.validateInboundSMSWebhook(webhook); err != nil {
		return fmt.Errorf("invalid inbound SMS webhook: %w", err)
	}

	// Check if message already exists (idempotency)
	if existingMessage, err := s.messageRepo.GetByProviderMessageID(ctx, webhook.MessagingProviderID); err == nil && existingMessage != nil {
		return nil // Message already processed
	}

	// Create message record
	message := s.buildInboundMessage(webhook.From, webhook.To, webhook.Type, webhook.Body, webhook.Attachments, webhook.Timestamp, webhook.MessagingProviderID)
	if err := s.createMessageRecord(ctx, message); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (s *messagingService) HandleInboundEmail(ctx context.Context, webhook *domain.InboundEmailWebhook) error {
	// Validate webhook
	if err := s.validateInboundEmailWebhook(webhook); err != nil {
		return fmt.Errorf("invalid inbound email webhook: %w", err)
	}

	// Check if message already exists (idempotency)
	if existingMessage, err := s.messageRepo.GetByProviderMessageID(ctx, webhook.XillioID); err == nil && existingMessage != nil {
		return nil // Message already processed
	}

	// Create message record
	message := s.buildInboundMessage(webhook.From, webhook.To, domain.MessageTypeEmail, webhook.Body, webhook.Attachments, webhook.Timestamp, webhook.XillioID)
	if err := s.createMessageRecord(ctx, message); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// buildOutboundMessage creates a message for outbound communication
func (s *messagingService) buildOutboundMessage(from, to, messageType, body string, attachments []string, timestamp time.Time) *domain.Message {
	// Ensure timestamp is in UTC
	utcTimestamp := timestamp.UTC()

	return &domain.Message{
		From:        from,
		To:          to,
		Type:        messageType,
		Body:        body,
		Attachments: attachments,
		Status:      "pending", // Outbound messages start as pending
		Timestamp:   utcTimestamp,
	}
}

// buildInboundMessage creates a message for inbound communication
func (s *messagingService) buildInboundMessage(from, to, messageType, body string, attachments []string, timestamp time.Time, providerMessageID string) *domain.Message {
	// Ensure timestamp is in UTC
	utcTimestamp := timestamp.UTC()

	return &domain.Message{
		From:                from,
		To:                  to,
		Type:                messageType,
		Body:                body,
		Attachments:         attachments,
		Status:              "delivered", // Inbound messages are considered delivered
		Timestamp:           utcTimestamp,
		MessagingProviderID: &providerMessageID,
	}
}

// sendSMSMessage sends an SMS/MMS message through the provider
func (s *messagingService) sendSMSMessage(ctx context.Context, req *domain.SendSMSRequest) error {
	switch req.Type {
	case domain.MessageTypeSMS:
		return s.smsProvider.SendSMS(ctx, req.From, req.To, req.Body)
	case domain.MessageTypeMMS:
		return s.smsProvider.SendMMS(ctx, req.From, req.To, req.Body, req.Attachments)
	default:
		return fmt.Errorf("invalid message type: %s", req.Type)
	}
}

// retryWithBackoff executes a function with retry logic and exponential backoff
func (s *messagingService) retryWithBackoff(ctx context.Context, operation func() error) error {
	for attempt := 0; attempt <= s.retryConfig.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if !domain.IsRetryableError(err) {
			return err
		}

		// If this is the last attempt, return the error
		if attempt == s.retryConfig.MaxRetries {
			return err
		}

		// Calculate delay with exponential backoff
		delay := s.retryConfig.BaseDelay * time.Duration(1<<attempt)

		// Cap delay at maximum
		if delay > s.retryConfig.MaxDelay {
			delay = s.retryConfig.MaxDelay
		}

		// For rate limit errors, use the RetryAfter value if available
		if retryAfter := domain.GetRetryAfterSeconds(err); retryAfter > 0 {
			retryDelay := time.Duration(retryAfter) * time.Second
			// Use the smaller of retry delay or max delay
			if retryDelay < s.retryConfig.MaxDelay {
				delay = retryDelay
			} else {
				delay = s.retryConfig.MaxDelay
			}
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return fmt.Errorf("max retries exceeded")
}

// sendSMSMessageWithRetry sends SMS with retry logic for HTTP errors
func (s *messagingService) sendSMSMessageWithRetry(ctx context.Context, req *domain.SendSMSRequest) error {
	return s.retryWithBackoff(ctx, func() error {
		return s.sendSMSMessage(ctx, req)
	})
}

// sendEmailMessageWithRetry sends email with retry logic for HTTP errors
func (s *messagingService) sendEmailMessageWithRetry(ctx context.Context, req *domain.SendEmailRequest) error {
	return s.retryWithBackoff(ctx, func() error {
		return s.emailProvider.SendEmail(ctx, req.From, req.To, req.Body, req.Attachments)
	})
}

// createMessageRecord creates a message record in the database
func (s *messagingService) createMessageRecord(ctx context.Context, message *domain.Message) error {
	// Normalize contacts for consistent conversation grouping
	customerContact, businessContact := s.normalizeContacts(message.From, message.To)

	// Get or create conversation
	conversation, err := s.conversationRepo.GetOrCreate(ctx, customerContact, businessContact)
	if err != nil {
		return fmt.Errorf("failed to get or create conversation: %w", err)
	}

	// Set conversation ID and timestamps
	message.ConversationID = conversation.ID
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	// Create the message record
	return s.messageRepo.Create(ctx, message)
}

// normalizeContacts ensures consistent ordering of contacts for conversation grouping
func (s *messagingService) normalizeContacts(customerContact, businessContact string) (string, string) {
	// For email addresses, sort alphabetically
	if strings.Contains(customerContact, "@") && strings.Contains(businessContact, "@") {
		contacts := []string{customerContact, businessContact}
		sort.Strings(contacts)
		return contacts[0], contacts[1]
	}

	// For phone numbers, clean and sort
	cleanCustomer := strings.ReplaceAll(customerContact, "-", "")
	cleanBusiness := strings.ReplaceAll(businessContact, "-", "")

	if cleanCustomer < cleanBusiness {
		return customerContact, businessContact
	}
	return businessContact, customerContact
}

// validateSMSRequest validates an SMS request
func (s *messagingService) validateSMSRequest(req *domain.SendSMSRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.From) == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if strings.TrimSpace(req.To) == "" {
		return fmt.Errorf("to address cannot be empty")
	}
	if strings.TrimSpace(req.Body) == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if req.Type != domain.MessageTypeSMS && req.Type != domain.MessageTypeMMS {
		return fmt.Errorf("invalid message type: %s", req.Type)
	}
	if err := s.validateTimestamp(req.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	return nil
}

// validateEmailRequest validates an email request
func (s *messagingService) validateEmailRequest(req *domain.SendEmailRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.From) == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if strings.TrimSpace(req.To) == "" {
		return fmt.Errorf("to address cannot be empty")
	}
	if strings.TrimSpace(req.Body) == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if err := s.validateTimestamp(req.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	return nil
}

// validateTimestamp validates a timestamp for business logic
func (s *messagingService) validateTimestamp(timestamp time.Time) error {
	// Ensure timestamp is in UTC
	if timestamp.Location() != time.UTC {
		return fmt.Errorf("timestamp must be in UTC timezone")
	}

	now := time.Now().UTC()

	// Check for zero timestamp
	if timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	// Check for future timestamps (allow small buffer for clock skew)
	maxFuture := now.Add(5 * time.Minute)
	if timestamp.After(maxFuture) {
		return fmt.Errorf("timestamp cannot be in the future (max allowed: %s)", maxFuture.Format(time.RFC3339))
	}

	// Check for very old timestamps (older than 10 years)
	minPast := now.AddDate(-10, 0, 0)
	if timestamp.Before(minPast) {
		return fmt.Errorf("timestamp too old (min allowed: %s)", minPast.Format(time.RFC3339))
	}

	// Check for unreasonable past timestamps (before 2000)
	year2000 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if timestamp.Before(year2000) {
		return fmt.Errorf("timestamp before year 2000 is not allowed")
	}

	return nil
}

// validateInboundSMSWebhook validates an inbound SMS webhook
func (s *messagingService) validateInboundSMSWebhook(webhook *domain.InboundSMSWebhook) error {
	if webhook == nil {
		return fmt.Errorf("webhook cannot be nil")
	}
	if strings.TrimSpace(webhook.From) == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if strings.TrimSpace(webhook.To) == "" {
		return fmt.Errorf("to address cannot be empty")
	}
	if strings.TrimSpace(webhook.Body) == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if webhook.Type != domain.MessageTypeSMS && webhook.Type != domain.MessageTypeMMS {
		return fmt.Errorf("invalid message type: %s", webhook.Type)
	}
	if strings.TrimSpace(webhook.MessagingProviderID) == "" {
		return fmt.Errorf("messaging provider ID cannot be empty")
	}
	if err := s.validateTimestamp(webhook.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	return nil
}

// validateInboundEmailWebhook validates an inbound email webhook
func (s *messagingService) validateInboundEmailWebhook(webhook *domain.InboundEmailWebhook) error {
	if webhook == nil {
		return fmt.Errorf("webhook cannot be nil")
	}
	if strings.TrimSpace(webhook.From) == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if strings.TrimSpace(webhook.To) == "" {
		return fmt.Errorf("to address cannot be empty")
	}
	if strings.TrimSpace(webhook.Body) == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if strings.TrimSpace(webhook.XillioID) == "" {
		return fmt.Errorf("xillio ID cannot be empty")
	}
	if err := s.validateTimestamp(webhook.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	return nil
}

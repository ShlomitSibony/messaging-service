package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Message types
const (
	MessageTypeSMS   = "sms"
	MessageTypeMMS   = "mms"
	MessageTypeEmail = "email"
)

// Message represents a message in the system
type Message struct {
	ID                  int       `json:"id" db:"id"`
	ConversationID      int       `json:"conversation_id" db:"conversation_id"`
	From                string    `json:"from" db:"from_address"`
	To                  string    `json:"to" db:"to_address"`
	Type                string    `json:"type" db:"message_type"`
	Body                string    `json:"body" db:"body"`
	Attachments         []string  `json:"attachments" db:"attachments"`
	Status              string    `json:"status" db:"status"`
	ErrorCode           *string   `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage        *string   `json:"error_message,omitempty" db:"error_message"`
	Timestamp           time.Time `json:"timestamp" db:"timestamp"`
	MessagingProviderID *string   `json:"messaging_provider_id,omitempty" db:"provider_message_id"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// Conversation represents a conversation between participants
type Conversation struct {
	ID              int       `json:"id" db:"id"`
	CustomerContact string    `json:"customer_contact" db:"customer_contact"`
	BusinessContact string    `json:"business_contact" db:"business_contact"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Messages        []Message `json:"messages,omitempty"`
}

// OutboundSMSRequest represents a request to send an SMS/MMS
type OutboundSMSRequest struct {
	From        string    `json:"from" binding:"required"`
	To          string    `json:"to" binding:"required"`
	Type        string    `json:"type" binding:"required,oneof=sms mms"`
	Body        string    `json:"body" binding:"required"`
	Attachments []string  `json:"attachments"`
	Timestamp   time.Time `json:"timestamp"`
}

// OutboundEmailRequest represents a request to send an email
type OutboundEmailRequest struct {
	From        string    `json:"from" binding:"required"`
	To          string    `json:"to" binding:"required"`
	Body        string    `json:"body" binding:"required"`
	Attachments []string  `json:"attachments"`
	Timestamp   time.Time `json:"timestamp"`
}

// InboundSMSWebhook represents an incoming SMS/MMS webhook
type InboundSMSWebhook struct {
	From                string    `json:"from" binding:"required"`
	To                  string    `json:"to" binding:"required"`
	Type                string    `json:"type" binding:"required,oneof=sms mms"`
	MessagingProviderID string    `json:"messaging_provider_id" binding:"required"`
	Body                string    `json:"body" binding:"required"`
	Attachments         []string  `json:"attachments"`
	Timestamp           time.Time `json:"timestamp,omitempty"`
}

// InboundEmailWebhook represents an incoming email webhook
type InboundEmailWebhook struct {
	From        string    `json:"from" binding:"required"`
	To          string    `json:"to" binding:"required"`
	XillioID    string    `json:"xillio_id" binding:"required"`
	Body        string    `json:"body" binding:"required"`
	Attachments []string  `json:"attachments"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

// API Response Types

// SendSMSRequest represents a request to send an SMS/MMS
type SendSMSRequest struct {
	From        string    `json:"from" binding:"required"`
	To          string    `json:"to" binding:"required"`
	Type        string    `json:"type" binding:"required,oneof=sms mms"`
	Body        string    `json:"body" binding:"required"`
	Attachments []string  `json:"attachments"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

// SendSMSResponse represents the response for sending an SMS/MMS
type SendSMSResponse struct {
	Message string `json:"message"`
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	From        string    `json:"from" binding:"required"`
	To          string    `json:"to" binding:"required"`
	Body        string    `json:"body" binding:"required"`
	Attachments []string  `json:"attachments"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

// SendEmailResponse represents the response for sending an email
type SendEmailResponse struct {
	Message string `json:"message"`
}

// WebhookResponse represents the response for webhook processing
type WebhookResponse struct {
	Message string `json:"message"`
}

// ConversationQuery represents query parameters for getting conversations
type ConversationQuery struct {
	BusinessEmail   string    `form:"business_email"` // Filter by business email
	BusinessPhone   string    `form:"business_phone"` // Filter by business phone
	Search          string    `form:"search"`
	From            time.Time `form:"from"`
	To              time.Time `form:"to"`
	MessageType     string    `form:"message_type"`
	Limit           int       `form:"limit,default=50"`
	Offset          int       `form:"offset,default=0"`
	SortBy          string    `form:"sort_by,default=updated_at"`
	SortOrder       string    `form:"sort_order,default=desc"`
	IncludeMessages bool      `form:"include_messages,default=false"`
}

// GetConversationsResponse represents the response for getting conversations
type GetConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	PerPage       int            `json:"per_page"`
	HasMore       bool           `json:"has_more"`
}

// GetConversationMessagesResponse represents the response for getting conversation messages
type GetConversationMessagesResponse struct {
	Messages []Message `json:"messages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ProviderError represents an error from a messaging provider with HTTP status code
type ProviderError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"` // seconds
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider error %d: %s", e.Code, e.Message)
}

// IsRetryableError checks if the error is retryable (429, 500, 502, 503, 504)
func IsRetryableError(err error) bool {
	if providerErr, ok := err.(*ProviderError); ok {
		return providerErr.Code == 429 || providerErr.Code == 500 ||
			providerErr.Code == 502 || providerErr.Code == 503 || providerErr.Code == 504
	}
	return false
}

// GetRetryAfterSeconds returns the retry after duration for rate limit errors
func GetRetryAfterSeconds(err error) int {
	if providerErr, ok := err.(*ProviderError); ok && providerErr.Code == 429 {
		return providerErr.RetryAfter
	}
	return 0
}

// Scan implements the sql.Scanner interface for Attachments
func (m *Message) Scan(value interface{}) error {
	if value == nil {
		m.Attachments = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &m.Attachments)
}

// Value implements the driver.Valuer interface for Attachments
func (m Message) Value() (driver.Value, error) {
	if m.Attachments == nil {
		return nil, nil
	}
	return json.Marshal(m.Attachments)
}

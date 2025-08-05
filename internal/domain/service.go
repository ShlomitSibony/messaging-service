package domain

import "context"

// MessagingService defines the interface for messaging operations
type MessagingService interface {
	SendSMS(ctx context.Context, req *SendSMSRequest) error
	SendEmail(ctx context.Context, req *SendEmailRequest) error
	HandleInboundSMS(ctx context.Context, webhook *InboundSMSWebhook) error
	HandleInboundEmail(ctx context.Context, webhook *InboundEmailWebhook) error
}

// ConversationService defines the interface for conversation operations
type ConversationService interface {
	GetConversations(ctx context.Context, query *ConversationQuery) (*GetConversationsResponse, error)
	GetConversationMessages(ctx context.Context, conversationID int) ([]Message, error)
}

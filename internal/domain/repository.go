package domain

import "context"

// ConversationRepository defines the interface for conversation data access
type ConversationRepository interface {
	Create(ctx context.Context, customerContact, businessContact string) (*Conversation, error)
	GetByID(ctx context.Context, id int) (*Conversation, error)
	GetByContacts(ctx context.Context, customerContact, businessContact string) (*Conversation, error)
	GetOrCreate(ctx context.Context, customerContact, businessContact string) (*Conversation, error)
	List(ctx context.Context, query *ConversationQuery) ([]Conversation, int, error)
}

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id int) (*Message, error)
	GetByConversationID(ctx context.Context, conversationID int) ([]Message, error)
	GetByProviderMessageID(ctx context.Context, providerMessageID string) (*Message, error)
	Update(ctx context.Context, message *Message) error
}

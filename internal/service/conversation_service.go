package service

import (
	"context"
	"fmt"
	"messaging-service/internal/domain"
	"strings"
)

type conversationService struct {
	conversationRepo domain.ConversationRepository
	messageRepo      domain.MessageRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(
	conversationRepo domain.ConversationRepository,
	messageRepo domain.MessageRepository,
) domain.ConversationService {
	return &conversationService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

// normalizeContacts ensures consistent ordering of contacts for conversation grouping
func (s *conversationService) normalizeContacts(customerContact, businessContact string) (string, string) {
	// For phone numbers, normalize by removing any formatting
	customer := strings.ReplaceAll(customerContact, "-", "")
	customer = strings.ReplaceAll(customer, " ", "")
	customer = strings.ReplaceAll(customer, "(", "")
	customer = strings.ReplaceAll(customer, ")", "")

	business := strings.ReplaceAll(businessContact, "-", "")
	business = strings.ReplaceAll(business, " ", "")
	business = strings.ReplaceAll(business, "(", "")
	business = strings.ReplaceAll(business, ")", "")

	// For email addresses, use as-is
	if strings.Contains(customer, "@") && strings.Contains(business, "@") {
		// For emails, sort alphabetically to ensure consistent ordering
		if customer < business {
			return customer, business
		}
		return business, customer
	}

	// For phone numbers, sort numerically
	if customer < business {
		return customer, business
	}
	return business, customer
}

func (s *conversationService) GetConversations(ctx context.Context, query *domain.ConversationQuery) (*domain.GetConversationsResponse, error) {
	// Set default values if not provided
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.SortBy == "" {
		query.SortBy = "updated_at"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	conversations, total, err := s.conversationRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Load messages for each conversation if requested
	if query.IncludeMessages {
		for i := range conversations {
			messages, err := s.messageRepo.GetByConversationID(ctx, conversations[i].ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get messages for conversation %d: %w", conversations[i].ID, err)
			}
			conversations[i].Messages = messages
		}
	}

	// Calculate pagination info
	page := (query.Offset / query.Limit) + 1
	hasMore := (query.Offset + query.Limit) < total

	return &domain.GetConversationsResponse{
		Conversations: conversations,
		Total:         total,
		Page:          page,
		PerPage:       query.Limit,
		HasMore:       hasMore,
	}, nil
}

func (s *conversationService) GetConversationMessages(ctx context.Context, conversationID int) ([]domain.Message, error) {
	// Verify conversation exists
	conversation, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if conversation == nil {
		return nil, fmt.Errorf("conversation not found: %d", conversationID)
	}

	// Get messages for the conversation
	messages, err := s.messageRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for conversation %d: %w", conversationID, err)
	}

	return messages, nil
}

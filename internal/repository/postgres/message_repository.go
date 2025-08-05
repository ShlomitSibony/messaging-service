package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"messaging-service/internal/domain"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type messageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (conversation_id, from_address, to_address, message_type, body, attachments, provider_message_id, status, timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	// Serialize attachments to JSON
	attachmentsJSON, err := json.Marshal(message.Attachments)
	if err != nil {
		return fmt.Errorf("failed to marshal attachments: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		message.ConversationID,
		message.From,
		message.To,
		message.Type,
		message.Body,
		attachmentsJSON,
		message.MessagingProviderID,
		message.Status,
		message.Timestamp,
		message.CreatedAt,
		message.UpdatedAt,
	).Scan(&message.ID)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (r *messageRepository) GetByID(ctx context.Context, id int) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, from_address, to_address, message_type, body, attachments, provider_message_id, status, error_code, error_message, timestamp, created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	var message domain.Message
	var attachmentsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConversationID,
		&message.From,
		&message.To,
		&message.Type,
		&message.Body,
		&attachmentsJSON,
		&message.MessagingProviderID,
		&message.Status,
		&message.ErrorCode,
		&message.ErrorMessage,
		&message.Timestamp,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	// Deserialize attachments from JSON
	if err := json.Unmarshal(attachmentsJSON, &message.Attachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
	}

	return &message, nil
}

func (r *messageRepository) GetByProviderMessageID(ctx context.Context, providerMessageID string) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, from_address, to_address, message_type, body, attachments, provider_message_id, status, error_code, error_message, timestamp, created_at, updated_at
		FROM messages
		WHERE provider_message_id = $1
	`

	var message domain.Message
	var attachmentsJSON []byte

	err := r.db.QueryRowContext(ctx, query, providerMessageID).Scan(
		&message.ID,
		&message.ConversationID,
		&message.From,
		&message.To,
		&message.Type,
		&message.Body,
		&attachmentsJSON,
		&message.MessagingProviderID,
		&message.Status,
		&message.ErrorCode,
		&message.ErrorMessage,
		&message.Timestamp,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by provider ID: %w", err)
	}

	// Deserialize attachments from JSON
	if err := json.Unmarshal(attachmentsJSON, &message.Attachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
	}

	return &message, nil
}

func (r *messageRepository) GetByConversationID(ctx context.Context, conversationID int) ([]domain.Message, error) {
	query := `
		SELECT id, conversation_id, from_address, to_address, message_type, body, attachments, provider_message_id, status, error_code, error_message, timestamp, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by conversation ID: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var message domain.Message
		var attachmentsJSON []byte

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.From,
			&message.To,
			&message.Type,
			&message.Body,
			&attachmentsJSON,
			&message.MessagingProviderID,
			&message.Status,
			&message.ErrorCode,
			&message.ErrorMessage,
			&message.Timestamp,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Deserialize attachments from JSON
		if err := json.Unmarshal(attachmentsJSON, &message.Attachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

func (r *messageRepository) Update(ctx context.Context, message *domain.Message) error {
	query := `
		UPDATE messages 
		SET status = $1, error_code = $2, error_message = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query,
		message.Status,
		message.ErrorCode,
		message.ErrorMessage,
		message.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

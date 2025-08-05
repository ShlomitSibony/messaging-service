package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"messaging-service/internal/domain"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type conversationRepository struct {
	db *sql.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sql.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	query := `
		INSERT INTO conversations (customer_contact, business_contact)
		VALUES ($1, $2)
		RETURNING id, customer_contact, business_contact, created_at, updated_at
	`

	var conv domain.Conversation
	err := r.db.QueryRowContext(ctx, query, customerContact, businessContact).Scan(
		&conv.ID,
		&conv.CustomerContact,
		&conv.BusinessContact,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &conv, nil
}

func (r *conversationRepository) GetByID(ctx context.Context, id int) (*domain.Conversation, error) {
	query := `
		SELECT id, customer_contact, business_contact, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	var conv domain.Conversation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.CustomerContact,
		&conv.BusinessContact,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get conversation by ID: %w", err)
	}

	return &conv, nil
}

func (r *conversationRepository) GetByContacts(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	query := `
		SELECT id, customer_contact, business_contact, created_at, updated_at
		FROM conversations
		WHERE (customer_contact = $1 AND business_contact = $2) OR (customer_contact = $2 AND business_contact = $1)
	`

	var conv domain.Conversation
	err := r.db.QueryRowContext(ctx, query, customerContact, businessContact).Scan(
		&conv.ID,
		&conv.CustomerContact,
		&conv.BusinessContact,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get conversation by contacts: %w", err)
	}

	return &conv, nil
}

func (r *conversationRepository) GetOrCreate(ctx context.Context, customerContact, businessContact string) (*domain.Conversation, error) {
	// Try to get existing conversation
	conv, err := r.GetByContacts(ctx, customerContact, businessContact)
	if err != nil {
		return nil, err
	}

	if conv != nil {
		return conv, nil
	}

	// Create new conversation if it doesn't exist
	return r.Create(ctx, customerContact, businessContact)
}

func (r *conversationRepository) List(ctx context.Context, query *domain.ConversationQuery) ([]domain.Conversation, int, error) {
	// Build the base query
	baseQuery := `
		SELECT id, customer_contact, business_contact, created_at, updated_at
		FROM conversations
		WHERE 1=1
	`

	// Build count query for pagination
	countQuery := `
		SELECT COUNT(*)
		FROM conversations
		WHERE 1=1
	`

	var args []interface{}
	var conditions []string
	argIndex := 1

	// Add filters

	if !query.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("updated_at >= $%d", argIndex))
		args = append(args, query.From)
		argIndex++
	}

	if !query.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("updated_at <= $%d", argIndex))
		args = append(args, query.To)
		argIndex++
	}

	// Add search functionality (search in both contacts)
	if query.Search != "" {
		searchCondition := fmt.Sprintf("(customer_contact ILIKE $%d OR business_contact ILIKE $%d)", argIndex, argIndex)
		conditions = append(conditions, searchCondition)
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	// Add business email filtering
	if query.BusinessEmail != "" {
		emailCondition := fmt.Sprintf("business_contact ILIKE $%d", argIndex)
		conditions = append(conditions, emailCondition)
		args = append(args, "%"+query.BusinessEmail+"%")
		argIndex++
	}

	// Add business phone filtering
	if query.BusinessPhone != "" {
		phoneCondition := fmt.Sprintf("business_contact ILIKE $%d", argIndex)
		conditions = append(conditions, phoneCondition)
		args = append(args, "%"+query.BusinessPhone+"%")
		argIndex++
	}

	// Add conditions to both queries
	for _, condition := range conditions {
		baseQuery += " AND " + condition
		countQuery += " AND " + condition
	}

	// Get total count for pagination
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Add sorting and pagination
	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	sortBy := "updated_at"
	if query.SortBy != "" {
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]bool{
			"id": true, "created_at": true, "updated_at": true,
			"customer_contact": true, "business_contact": true,
		}
		if validSortFields[query.SortBy] {
			sortBy = query.SortBy
		}
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, query.Limit, query.Offset)

	// Execute the query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		var conv domain.Conversation
		err := rows.Scan(
			&conv.ID,
			&conv.CustomerContact,
			&conv.BusinessContact,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating conversations: %w", err)
	}

	return conversations, total, nil
}

package container

import (
	"database/sql"

	"messaging-service/internal/config"
	"messaging-service/internal/domain"
	"messaging-service/internal/handler"
	"messaging-service/internal/provider"
	"messaging-service/internal/repository/postgres"
	"messaging-service/internal/service"
)

// Container holds all application dependencies
type Container struct {
	Config              *config.Config
	DB                  *sql.DB
	ConversationRepo    domain.ConversationRepository
	MessageRepo         domain.MessageRepository
	SMSProvider         domain.SMSProvider
	EmailProvider       domain.EmailProvider
	MessagingService    domain.MessagingService
	ConversationService domain.ConversationService
	MessagingHandler    *handler.MessagingHandler
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, db *sql.DB) *Container {
	container := &Container{
		Config: cfg,
		DB:     db,
	}

	// Initialize repositories
	container.ConversationRepo = postgres.NewConversationRepository(db)
	container.MessageRepo = postgres.NewMessageRepository(db)

	// Initialize providers
	container.SMSProvider = provider.NewMockSMSProvider()
	container.EmailProvider = provider.NewEmailProvider(
		provider.EmailProviderType(container.Config.Providers.EmailProviderType),
		container.Config.Providers.EmailProviderConfig,
	)

	// Initialize services
	container.MessagingService = service.NewMessagingService(
		container.ConversationRepo,
		container.MessageRepo,
		container.SMSProvider,
		container.EmailProvider,
	)
	container.ConversationService = service.NewConversationService(
		container.ConversationRepo,
		container.MessageRepo,
	)

	// Initialize handlers
	container.MessagingHandler = handler.NewMessagingHandler(
		container.MessagingService,
		container.ConversationService,
	)

	return container
}

// Close closes all resources in the container
func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

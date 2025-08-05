package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"messaging-service/internal/domain"
	"messaging-service/internal/handler"
	"messaging-service/internal/logger"
	"messaging-service/internal/middleware"
	"messaging-service/internal/provider"
	"messaging-service/internal/repository/postgres"
	"messaging-service/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	db                  *sql.DB
	conversationRepo    domain.ConversationRepository
	messageRepo         domain.MessageRepository
	messagingService    domain.MessagingService
	conversationService domain.ConversationService
	handler             *handler.MessagingHandler
	router              *gin.Engine
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Connect to test database
	db, err := sql.Open("postgres", "host=localhost port=5432 dbname=messaging_service user=messaging_user password=messaging_password sslmode=disable")
	require.NoError(t, err)

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	// Clear test data
	_, err = db.Exec("DELETE FROM messages")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM conversations")
	require.NoError(t, err)

	// Initialize repositories
	conversationRepo := postgres.NewConversationRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	// Initialize providers
	smsProvider := provider.NewMockSMSProvider()
	emailProvider := provider.NewMockEmailProvider()

	// Initialize services
	messagingService := service.NewMessagingService(conversationRepo, messageRepo, smsProvider, emailProvider)
	conversationService := service.NewConversationService(conversationRepo, messageRepo)

	// Initialize handler
	messagingHandler := handler.NewMessagingHandler(messagingService, conversationService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add routes with middleware
	api := router.Group("/api")
	api.Use(
		middleware.RequestIDMiddleware(),
		middleware.LoggingMiddleware(logger.Get()),
		middleware.MetricsMiddleware(),
	)
	{
		messages := api.Group("/messages")
		{
			messages.POST("/message", messagingHandler.SendSMS)
			messages.POST("/email", messagingHandler.SendEmail)
		}

		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/message", messagingHandler.HandleInboundSMS)
			webhooks.POST("/email", messagingHandler.HandleInboundEmail)
		}

		conversations := api.Group("/conversations")
		{
			conversations.GET("", messagingHandler.GetConversations)
			conversations.GET("/:id/messages", messagingHandler.GetConversationMessages)
		}
	}

	return &IntegrationTestSuite{
		db:                  db,
		conversationRepo:    conversationRepo,
		messageRepo:         messageRepo,
		messagingService:    messagingService,
		conversationService: conversationService,
		handler:             messagingHandler,
		router:              router,
	}
}

func (suite *IntegrationTestSuite) cleanup() {
	suite.db.Close()
}

func TestIntegration_SendSMSAndRetrieveConversation(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Send SMS
	smsRequest := domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "First message",
	}

	err := suite.messagingService.SendSMS(context.Background(), &smsRequest)
	assert.NoError(t, err)

	// Send another message to create a conversation
	smsRequest2 := domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "Second message",
	}

	err = suite.messagingService.SendSMS(context.Background(), &smsRequest2)
	assert.NoError(t, err)

	// Get conversations
	conversations, err := suite.conversationService.GetConversations(context.Background(), &domain.ConversationQuery{
		Limit:           10,
		IncludeMessages: true,
	})
	assert.NoError(t, err)
	assert.Len(t, conversations.Conversations, 1)

	conversation := conversations.Conversations[0]
	assert.Equal(t, "+12016661234", conversation.CustomerContact)
	assert.Equal(t, "+18045551234", conversation.BusinessContact)
	assert.Len(t, conversation.Messages, 2)

	// Verify messages
	assert.Equal(t, "First message", conversation.Messages[0].Body)
	assert.Equal(t, "sms", conversation.Messages[0].Type)
	assert.Equal(t, "+12016661234", conversation.Messages[0].From)
	assert.Equal(t, "+18045551234", conversation.Messages[0].To)

	assert.Equal(t, "Second message", conversation.Messages[1].Body)
	assert.Equal(t, "sms", conversation.Messages[1].Type)
	assert.Equal(t, "+12016661234", conversation.Messages[1].From)
	assert.Equal(t, "+18045551234", conversation.Messages[1].To)
}

func TestIntegration_SendEmailAndRetrieveConversation(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Send email
	emailRequest := domain.SendEmailRequest{
		Timestamp:   time.Now().UTC(),
		From:        "user@usehatchapp.com",
		To:          "contact@gmail.com",
		Body:        "Hello! This is a test email message with <b>HTML</b> formatting.",
		Attachments: []string{"https://example.com/document.pdf"},
	}

	err := suite.messagingService.SendEmail(context.Background(), &emailRequest)
	assert.NoError(t, err)

	// Get conversations
	conversations, err := suite.conversationService.GetConversations(context.Background(), &domain.ConversationQuery{
		Limit:           10,
		IncludeMessages: true,
	})
	assert.NoError(t, err)
	assert.Len(t, conversations.Conversations, 1)

	conversation := conversations.Conversations[0]
	assert.Equal(t, "contact@gmail.com", conversation.CustomerContact)
	assert.Equal(t, "user@usehatchapp.com", conversation.BusinessContact)
	assert.Len(t, conversation.Messages, 1)

	// Verify message
	message := conversation.Messages[0]
	assert.Equal(t, "Hello! This is a test email message with <b>HTML</b> formatting.", message.Body)
	assert.Equal(t, "email", message.Type)
	assert.Equal(t, "user@usehatchapp.com", message.From)
	assert.Equal(t, "contact@gmail.com", message.To)
	assert.Equal(t, []string{"https://example.com/document.pdf"}, message.Attachments)
}

func TestIntegration_HandleInboundSMSWebhook(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Send initial SMS
	smsRequest := domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "Hello! This is a test SMS message.",
	}

	err := suite.messagingService.SendSMS(context.Background(), &smsRequest)
	assert.NoError(t, err)

	// Handle inbound SMS webhook
	webhook := domain.InboundSMSWebhook{
		Timestamp:           time.Now().UTC(),
		From:                "+18045551234",
		To:                  "+12016661234",
		Type:                "sms",
		MessagingProviderID: "message-1",
		Body:                "This is an incoming SMS message",
	}

	err = suite.messagingService.HandleInboundSMS(context.Background(), &webhook)
	assert.NoError(t, err)

	// Get conversations
	conversations, err := suite.conversationService.GetConversations(context.Background(), &domain.ConversationQuery{
		Limit:           10,
		IncludeMessages: true,
	})
	assert.NoError(t, err)
	assert.Len(t, conversations.Conversations, 1)

	conversation := conversations.Conversations[0]
	assert.Len(t, conversation.Messages, 2)

	// Verify outbound message
	outboundMessage := conversation.Messages[0]
	assert.Equal(t, "Hello! This is a test SMS message.", outboundMessage.Body)
	assert.Equal(t, "sms", outboundMessage.Type)
	assert.Equal(t, "+12016661234", outboundMessage.From)
	assert.Equal(t, "+18045551234", outboundMessage.To)

	// Verify inbound message
	inboundMessage := conversation.Messages[1]
	assert.Equal(t, "This is an incoming SMS message", inboundMessage.Body)
	assert.Equal(t, "sms", inboundMessage.Type)
	assert.Equal(t, "+18045551234", inboundMessage.From)
	assert.Equal(t, "+12016661234", inboundMessage.To)
	assert.Equal(t, "message-1", *inboundMessage.MessagingProviderID)
}

func TestIntegration_HandleInboundEmailWebhook(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Send initial email
	emailRequest := domain.SendEmailRequest{
		Timestamp:   time.Now().UTC(),
		From:        "user@usehatchapp.com",
		To:          "contact@gmail.com",
		Body:        "Hello! This is a test email message with <b>HTML</b> formatting.",
		Attachments: []string{"https://example.com/document.pdf"},
	}

	err := suite.messagingService.SendEmail(context.Background(), &emailRequest)
	assert.NoError(t, err)

	// Handle inbound email webhook
	webhook := domain.InboundEmailWebhook{
		Timestamp:   time.Now().UTC(),
		From:        "contact@gmail.com",
		To:          "user@usehatchapp.com",
		XillioID:    "message-3",
		Body:        "<html><body>This is an incoming email with <b>HTML</b> content</body></html>",
		Attachments: []string{"https://example.com/received-document.pdf"},
	}

	err = suite.messagingService.HandleInboundEmail(context.Background(), &webhook)
	assert.NoError(t, err)

	// Get conversations
	conversations, err := suite.conversationService.GetConversations(context.Background(), &domain.ConversationQuery{
		Limit:           10,
		IncludeMessages: true,
	})
	assert.NoError(t, err)
	assert.Len(t, conversations.Conversations, 1)

	conversation := conversations.Conversations[0]
	assert.Len(t, conversation.Messages, 2)

	// Verify outbound message
	outboundMessage := conversation.Messages[0]
	assert.Equal(t, "Hello! This is a test email message with <b>HTML</b> formatting.", outboundMessage.Body)
	assert.Equal(t, "email", outboundMessage.Type)
	assert.Equal(t, "user@usehatchapp.com", outboundMessage.From)
	assert.Equal(t, "contact@gmail.com", outboundMessage.To)

	// Verify inbound message
	inboundMessage := conversation.Messages[1]
	assert.Equal(t, "<html><body>This is an incoming email with <b>HTML</b> content</body></html>", inboundMessage.Body)
	assert.Equal(t, "email", inboundMessage.Type)
	assert.Equal(t, "contact@gmail.com", inboundMessage.From)
	assert.Equal(t, "user@usehatchapp.com", inboundMessage.To)
	assert.Equal(t, "message-3", *inboundMessage.MessagingProviderID)
	assert.Equal(t, []string{"https://example.com/received-document.pdf"}, inboundMessage.Attachments)
}

func TestIntegration_HTTPEndpoints(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Test SMS endpoint
	smsRequest := domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      "+12016661234",
		To:        "+18045551234",
		Type:      "sms",
		Body:      "HTTP test SMS",
	}

	jsonData, _ := json.Marshal(smsRequest)
	req, _ := http.NewRequest("POST", "/api/messages/message", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test email endpoint
	emailRequest := domain.SendEmailRequest{
		Timestamp: time.Now().UTC(),
		From:      "user@usehatchapp.com",
		To:        "contact@gmail.com",
		Body:      "HTTP test email",
	}

	jsonData, _ = json.Marshal(emailRequest)
	req, _ = http.NewRequest("POST", "/api/messages/email", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test conversations endpoint with query parameter
	req, _ = http.NewRequest("GET", "/api/conversations?business_phone=+12016661234", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "conversations")
}

func TestIntegration_ConversationGrouping(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Send messages between the same participants
	participant1 := "+12016661234"
	participant2 := "+18045551234"

	// Send SMS
	smsRequest := domain.SendSMSRequest{
		Timestamp: time.Now().UTC(),
		From:      participant1,
		To:        participant2,
		Type:      "sms",
		Body:      "First message",
	}
	err := suite.messagingService.SendSMS(context.Background(), &smsRequest)
	assert.NoError(t, err)

	// Send email
	emailRequest := domain.SendEmailRequest{
		Timestamp: time.Now().UTC(),
		From:      participant2,
		To:        participant1,
		Body:      "Reply via email",
	}
	err = suite.messagingService.SendEmail(context.Background(), &emailRequest)
	assert.NoError(t, err)

	// Verify only one conversation exists with both messages
	conversations, err := suite.conversationService.GetConversations(context.Background(), &domain.ConversationQuery{
		Limit:           10,
		IncludeMessages: true,
	})
	assert.NoError(t, err)
	assert.Len(t, conversations.Conversations, 1)

	conversation := conversations.Conversations[0]
	assert.Len(t, conversation.Messages, 2)

	// Verify messages are in chronological order
	assert.True(t, conversation.Messages[0].Timestamp.Before(conversation.Messages[1].Timestamp))
}

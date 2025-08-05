package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"messaging-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// MessagingHandler handles HTTP requests for messaging operations
type MessagingHandler struct {
	messagingService    domain.MessagingService
	conversationService domain.ConversationService
}

// NewMessagingHandler creates a new messaging handler
func NewMessagingHandler(messagingService domain.MessagingService, conversationService domain.ConversationService) *MessagingHandler {
	return &MessagingHandler{
		messagingService:    messagingService,
		conversationService: conversationService,
	}
}

// SendSMS godoc
// @Summary Send message
// @Description Send an SMS or MMS message to a recipient
// @Tags messages
// @Accept json
// @Produce json
// @Param message body domain.SendSMSRequest true "Message details"
// @Success 200 {object} domain.SendSMSResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /messages/message [post]
func (h *MessagingHandler) SendSMS(c *gin.Context) {
	var req domain.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set timestamp if not provided (always in UTC)
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	if err := h.messagingService.SendSMS(c.Request.Context(), &req); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to send SMS", err)
		return
	}

	c.JSON(http.StatusOK, domain.SendSMSResponse{Message: "Message sent successfully"})
}

// SendEmail godoc
// @Summary Send email message
// @Description Send an email message to a recipient
// @Tags messages
// @Accept json
// @Produce json
// @Param message body domain.SendEmailRequest true "Email message details"
// @Success 200 {object} domain.SendEmailResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /messages/email [post]
func (h *MessagingHandler) SendEmail(c *gin.Context) {
	var req domain.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set timestamp if not provided (always in UTC)
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	if err := h.messagingService.SendEmail(c.Request.Context(), &req); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to send email", err)
		return
	}

	c.JSON(http.StatusOK, domain.SendEmailResponse{Message: "Email sent successfully"})
}

// HandleInboundSMS godoc
// @Summary Handle incoming message webhook
// @Description Process incoming SMS and MMS messages from external providers
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body domain.InboundSMSWebhook true "Incoming message webhook data"
// @Success 200 {object} domain.WebhookResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /webhooks/message [post]
func (h *MessagingHandler) HandleInboundSMS(c *gin.Context) {
	var webhook domain.InboundSMSWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid webhook body", err)
		return
	}

	// Set timestamp if not provided
	if webhook.Timestamp.IsZero() {
		webhook.Timestamp = time.Now()
	}

	if err := h.messagingService.HandleInboundSMS(c.Request.Context(), &webhook); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to process inbound SMS", err)
		return
	}

	c.JSON(http.StatusOK, domain.WebhookResponse{Message: "Inbound message processed successfully"})
}

// HandleInboundEmail godoc
// @Summary Handle incoming email webhook
// @Description Process incoming email messages from external providers
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body domain.InboundEmailWebhook true "Incoming email webhook data"
// @Success 200 {object} domain.WebhookResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /webhooks/email [post]
func (h *MessagingHandler) HandleInboundEmail(c *gin.Context) {
	var webhook domain.InboundEmailWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid webhook body", err)
		return
	}

	// Set timestamp if not provided
	if webhook.Timestamp.IsZero() {
		webhook.Timestamp = time.Now()
	}

	if err := h.messagingService.HandleInboundEmail(c.Request.Context(), &webhook); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to process inbound email", err)
		return
	}

	c.JSON(http.StatusOK, domain.WebhookResponse{Message: "Inbound email processed successfully"})
}

// handleOutboundWebhook is a generic handler for outbound webhooks
func (h *MessagingHandler) handleOutboundWebhook(c *gin.Context, webhookType string, processFunc func(context.Context) error) {
	// Set timestamp if not provided
	if err := h.setTimestampIfZero(c); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid webhook body", err)
		return
	}

	if err := processFunc(c.Request.Context()); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to process outbound %s status", webhookType), err)
		return
	}

	c.JSON(http.StatusOK, domain.WebhookResponse{Message: fmt.Sprintf("Outbound %s status processed successfully", webhookType)})
}

// setTimestampIfZero sets timestamp to current time if it's zero
func (h *MessagingHandler) setTimestampIfZero(c *gin.Context) error {
	// This is a simplified version - in practice, you'd need to handle different webhook types
	// For now, we'll keep the individual handlers but extract common logic
	return nil
}

// GetConversations godoc
// @Summary Get conversations with filtering and pagination
// @Description Retrieve conversations with optional filtering, search, and pagination. At least one query parameter is required for performance reasons.
// @Tags conversations
// @Accept json
// @Produce json
// @Param business_email query string false "Filter by business email"
// @Param business_phone query string false "Filter by business phone"
// @Param search query string false "Search in customer or business contacts"
// @Param from query string false "Filter conversations updated from date (RFC3339)"
// @Param to query string false "Filter conversations updated to date (RFC3339)"
// @Param message_type query string false "Filter by message type (sms, mms, email)"
// @Param limit query int false "Number of conversations per page (default: 50, max: 100)"
// @Param offset query int false "Number of conversations to skip (default: 0)"
// @Param sort_by query string false "Sort field (id, created_at, updated_at)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param include_messages query bool false "Include messages in response (default: false)"
// @Success 200 {object} domain.GetConversationsResponse
// @Failure 400 {object} domain.ErrorResponse "Invalid query parameters or no query parameters provided"
// @Failure 500 {object} domain.ErrorResponse
// @Router /conversations [get]
func (h *MessagingHandler) GetConversations(c *gin.Context) {
	var query domain.ConversationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Validate that at least one query parameter is provided for performance reasons
	if query.BusinessEmail == "" && query.BusinessPhone == "" && query.Search == "" &&
		query.From.IsZero() && query.To.IsZero() && query.MessageType == "" {
		h.sendErrorResponse(c, http.StatusBadRequest, "At least one query parameter is required (business_email, business_phone, search, from, to, or message_type)", nil)
		return
	}

	// Validate and sanitize parameters
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	// Parse date parameters if provided
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			query.From = from
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			query.To = to
		}
	}

	response, err := h.conversationService.GetConversations(c.Request.Context(), &query)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to get conversations", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetConversationMessages godoc
// @Summary Get messages for a conversation
// @Description Retrieve all messages for a specific conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} domain.GetConversationMessagesResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /conversations/{id}/messages [get]
func (h *MessagingHandler) GetConversationMessages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err)
		return
	}

	messages, err := h.conversationService.GetConversationMessages(c.Request.Context(), id)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to get messages", err)
		return
	}

	c.JSON(http.StatusOK, domain.GetConversationMessagesResponse{Messages: messages})
}

// sendErrorResponse sends a consistent error response
func (h *MessagingHandler) sendErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = message + ": " + err.Error()
	}
	c.JSON(statusCode, domain.ErrorResponse{Error: errorMsg})
}

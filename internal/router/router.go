package router

import (
	"net/http"
	"time"

	"messaging-service/internal/handler"
	"messaging-service/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Router represents the HTTP router configuration
type Router struct {
	engine *gin.Engine
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	router := &Router{
		engine: gin.Default(),
	}

	return router
}

// SetupRoutes configures all routes with the given handler
func (r *Router) SetupRoutes(messagingHandler *handler.MessagingHandler, logger *zap.Logger) {
	// Health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Swagger documentation endpoint
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes with middleware
	api := r.engine.Group("/api")
	api.Use(
		middleware.RequestIDMiddleware(),
		middleware.LoggingMiddleware(logger),
		middleware.MetricsMiddleware(),
	)
	{
		// Message endpoints
		messages := api.Group("/messages")
		{
			messages.POST("/message", messagingHandler.SendSMS)
			messages.POST("/email", messagingHandler.SendEmail)
		}

		// Webhook endpoints
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/message", messagingHandler.HandleInboundSMS)
			webhooks.POST("/email", messagingHandler.HandleInboundEmail)
		}

		// Conversation endpoints
		conversations := api.Group("/conversations")
		{
			conversations.GET("", messagingHandler.GetConversations)
			conversations.GET("/:id/messages", messagingHandler.GetConversationMessages)
		}
	}
}

// GetEngine returns the underlying gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

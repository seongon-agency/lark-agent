package main

import (
	"context"
	"start-feishubot/handlers"
	"start-feishubot/initialization"
	"start-feishubot/logger"

	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/pflag"
	"start-feishubot/services/openai"
)

func main() {
	logger.Info("========================================")
	logger.Info("Starting Feishu Bot with ENHANCED LOGGING")
	logger.Info("========================================")

	initialization.InitRoleList()
	pflag.Parse()
	config := initialization.GetConfig()

	logger.Info("Configuration loaded")
	logger.Info("Verification Token:", config.FeishuAppVerificationToken)
	logger.Info("Encrypt Key:", config.FeishuAppEncryptKey)

	initialization.LoadLarkClient(*config)
	gpt := openai.NewChatGPT(*config)
	handlers.InitHandlers(gpt, *config)

	logger.Info("Handlers initialized successfully")

	eventHandler := dispatcher.NewEventDispatcher(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey).
		OnP2MessageReceiveV1(handlers.Handler).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			logger.Debugf("Received request %v", event.RequestURI)
			return handlers.ReadHandler(ctx, event)
		})

	logger.Info("Card webhook verification token:", config.FeishuAppVerificationToken)
	logger.Info("Card webhook encrypt key length:", len(config.FeishuAppEncryptKey))

	r := gin.Default()

	// Add recovery middleware with logging
	r.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		logger.Error("PANIC RECOVERED:", err)
		c.JSON(500, gin.H{"error": "internal server error"})
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Test endpoint to verify deployment
	r.GET("/test-card-logging", func(c *gin.Context) {
		logger.Info("Test endpoint called - logging is working!")
		c.JSON(200, gin.H{
			"message": "Logging enabled",
			"version": "enhanced-logging-v2",
			"timestamp": "2025-10-27",
		})
	})
	r.POST("/webhook/event",
		sdkginext.NewEventHandlerFunc(eventHandler))

	// Card webhook - Try using EVENT handler which works
	logger.Info("Registering card webhook handler...")
	logger.Info("WORKAROUND: Using event handler for card webhook since card handler has signature issues")

	// Create a minimal event dispatcher that just handles the challenge
	cardEventHandler := dispatcher.NewEventDispatcher(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey)

	r.POST("/webhook/card", func(c *gin.Context) {
		logger.Info("========== CARD WEBHOOK RECEIVED ==========")

		// First try event handler (which works) for challenge verification
		sdkginext.NewEventHandlerFunc(cardEventHandler)(c)

		logger.Info("Event handler (used for card) status:", c.Writer.Status())

		// If it succeeded (200), challenge was handled
		// If card actions come through later, we'll need to handle them differently
	})

	logger.Info("All routes registered successfully")
	logger.Info("Server starting...")

	if err := initialization.StartServer(*config, r); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

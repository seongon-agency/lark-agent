package main

import (
	"bytes"
	"context"
	"io"
	"start-feishubot/handlers"
	"start-feishubot/initialization"
	"start-feishubot/logger"

	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
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

	cardHandler := larkcard.NewCardActionHandler(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey,
		handlers.CardHandler())

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

	// Card webhook with logging wrapper
	r.POST("/webhook/card", func(c *gin.Context) {
		logger.Info("========== CARD WEBHOOK REQUEST ==========")

		// Log the raw request for debugging
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read body:", err)
			c.JSON(500, gin.H{"error": "failed to read body"})
			return
		}

		logger.Info("Body length:", len(bodyBytes))
		logger.Info("Raw body:", string(bodyBytes))

		// Restore body for SDK handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Let SDK handler process (it will decrypt and handle challenge)
		logger.Info("Passing to SDK handler for decryption...")
		sdkginext.NewCardActionHandlerFunc(cardHandler)(c)

		logger.Info("SDK handler completed, status:", c.Writer.Status())
	})

	logger.Info("All routes registered successfully")
	logger.Info("Server starting...")

	if err := initialization.StartServer(*config, r); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"start-feishubot/handlers"
	"start-feishubot/initialization"
	"start-feishubot/logger"

	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/pflag"
	"start-feishubot/services/openai"
)

// decryptCardWebhook decrypts the encrypted card webhook body
func decryptCardWebhook(bodyBytes []byte, encryptKey, verificationToken string) ([]byte, error) {
	// Parse the encrypted body
	var encryptedBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &encryptedBody); err != nil {
		return nil, err
	}

	// Check if it's encrypted
	if encryptStr, ok := encryptedBody["encrypt"].(string); ok {
		// Use the SDK's crypto helper to decrypt
		decrypted, err := httpserverext.Decrypt(encryptStr, encryptKey)
		if err != nil {
			return nil, err
		}
		return []byte(decrypted), nil
	}

	// Not encrypted, return as-is
	return bodyBytes, nil
}

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

	// Card webhook - Hybrid approach: use event handler for verification, custom handler for actions
	logger.Info("Registering card webhook handler...")
	logger.Info("WORKAROUND: Using hybrid handler since SDK card handler has signature issues")

	r.POST("/webhook/card", func(c *gin.Context) {
		logger.Info("========== CARD WEBHOOK RECEIVED ==========")

		// Read body once
		bodyBytes, err := c.GetRawData()
		if err != nil {
			logger.Error("Failed to read body:", err)
			c.JSON(500, gin.H{"error": "failed to read body"})
			return
		}

		logger.Debug("Raw body:", string(bodyBytes))

		// Try to decrypt using the event crypto package (which works)
		decryptedBody, err := decryptCardWebhook(bodyBytes, config.FeishuAppEncryptKey, config.FeishuAppVerificationToken)
		if err != nil {
			logger.Error("Failed to decrypt:", err)
			c.JSON(500, gin.H{"error": "decryption failed"})
			return
		}

		logger.Info("Successfully decrypted card webhook")
		logger.Debug("Decrypted body:", string(decryptedBody))

		// Check if it's a challenge
		var challengeBody map[string]interface{}
		if err := json.Unmarshal(decryptedBody, &challengeBody); err == nil {
			if challenge, ok := challengeBody["challenge"].(string); ok {
				logger.Info("✓ Challenge verified:", challenge)
				c.JSON(200, gin.H{"challenge": challenge})
				return
			}
		}

		// It's a card action - parse and handle it
		var cardAction larkcard.CardAction
		if err := json.Unmarshal(decryptedBody, &cardAction); err != nil {
			logger.Error("Failed to parse card action:", err)
			c.JSON(500, gin.H{"error": "invalid card action"})
			return
		}

		logger.Info("Processing card action...")
		result, err := handlers.CardHandler()(context.Background(), &cardAction)
		if err != nil {
			logger.Error("Card handler error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info("✓ Card action processed successfully")
		c.JSON(200, result)
	})

	logger.Info("All routes registered successfully")
	logger.Info("Server starting...")

	if err := initialization.StartServer(*config, r); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

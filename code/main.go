package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	initialization.InitRoleList()
	pflag.Parse()
	config := initialization.GetConfig()
	initialization.LoadLarkClient(*config)
	gpt := openai.NewChatGPT(*config)
	handlers.InitHandlers(gpt, *config)

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

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/webhook/event",
		sdkginext.NewEventHandlerFunc(eventHandler))

	// Manual challenge handler for card webhook
	r.POST("/webhook/card", func(c *gin.Context) {
		// Read the body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read card webhook body:", err)
			c.JSON(500, gin.H{"error": "failed to read body"})
			return
		}

		// Parse as JSON to check for challenge
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			logger.Error("Failed to parse card webhook JSON:", err)
			c.JSON(500, gin.H{"error": "invalid json"})
			return
		}

		// Check if this is a challenge request
		if challenge, ok := body["challenge"].(string); ok {
			logger.Info("Received challenge request for card webhook:", challenge)
			c.JSON(200, gin.H{"challenge": challenge})
			return
		}

		// Not a challenge - restore body and pass to SDK handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		sdkginext.NewCardActionHandlerFunc(cardHandler)(c)
	})

	if err := initialization.StartServer(*config, r); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

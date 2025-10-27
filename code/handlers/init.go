package handlers

import (
	"context"
	"start-feishubot/logger"

	"start-feishubot/initialization"
	"start-feishubot/services/openai"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type MessageHandlerInterface interface {
	msgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error
	cardHandler(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error)
}

type HandlerType string

const (
	GroupHandler = "group"
	UserHandler  = "personal"
)

// handlers - Handler for all message types
var handlers MessageHandlerInterface

func InitHandlers(gpt *openai.ChatGPT, config initialization.Config) {
	handlers = NewMessageHandler(gpt, config)
}

func Handler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	return handlers.msgReceivedHandler(ctx, event)
}

func ReadHandler(ctx context.Context, event *larkim.P2MessageReadV1) error {
	readerId := event.Event.Reader.ReaderId.OpenId
	//fmt.Printf("msg is read by : %v \n", *readerId)
	logger.Debugf("msg is read by : %v \n", *readerId)

	return nil
}

func CardHandler() func(ctx context.Context,
	cardAction *larkcard.CardAction) (interface{}, error) {
	return func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("❌ Card handler PANIC:", err)
			}
		}()

		logger.Info("✓ CardHandler called - SDK decryption successful")

		if handlers == nil {
			logger.Error("❌ Handlers not initialized!")
			return nil, nil
		}

		if cardAction == nil {
			logger.Error("❌ cardAction is nil!")
			return nil, nil
		}

		logger.Info("Card action type:", cardAction.Action)
		logger.Debugf("Full card action: %+v", cardAction)

		result, err := handlers.cardHandler(ctx, cardAction)
		if err != nil {
			logger.Error("❌ Card handler error:", err)
		} else {
			logger.Info("✓ Card handler completed successfully")
		}
		return result, err
	}
}

func judgeCardType(cardAction *larkcard.CardAction) HandlerType {
	actionValue := cardAction.Action.Value
	chatType := actionValue["chatType"]
	//fmt.Printf("chatType: %v", chatType)
	if chatType == "group" {
		return GroupHandler
	}
	if chatType == "personal" {
		return UserHandler
	}
	return "otherChat"
}

func judgeChatType(event *larkim.P2MessageReceiveV1) HandlerType {
	chatType := event.Event.Message.ChatType
	if *chatType == "group" {
		return GroupHandler
	}
	if *chatType == "p2p" {
		return UserHandler
	}
	return "otherChat"
}

package handlers

import (
	"context"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"start-feishubot/logger"
	"start-feishubot/services"
)

func NewClearCardHandler(cardMsg CardMsg, m MessageHandler) CardHandlerFunc {
	return func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		if cardMsg.Kind == ClearCardKind {
			newCard, err, done := CommonProcessClearCache(cardMsg, m.sessionCache)
			if done {
				return newCard, err
			}
			return nil, nil
		}
		return nil, ErrNextHandler
	}
}

func CommonProcessClearCache(cardMsg CardMsg, session services.SessionServiceCacheInterface) (
	interface{}, error, bool) {
	logger.Debugf("card msg value %v", cardMsg.Value)
	if cardMsg.Value == "1" {
		session.Clear(cardMsg.SessionId)
		newCard, _ := newSendCard(
			withHeader("🆑 Bot Reminder", larkcard.TemplateGrey),
			withMainMd("Context information for this topic has been deleted"),
			withNote("We can start a brand new topic, feel free to continue chatting with me"),
		)
		logger.Debugf("session %v", newCard)
		return newCard, nil, true
	}
	if cardMsg.Value == "0" {
		newCard, _ := newSendCard(
			withHeader("🆑 Bot Reminder", larkcard.TemplateGreen),
			withMainMd("Context information for this topic is still retained"),
			withNote("We can continue discussing this topic, looking forward to chatting with you. If you have other questions or topics you'd like to discuss, please let me know"),
		)
		return newCard, nil, true
	}
	return nil, nil, false
}

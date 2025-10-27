package handlers

import (
	"context"

	"start-feishubot/services"
	"start-feishubot/services/openai"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
)

// AIModeChooseKind is the kind of card action for choosing AI mode
func NewAIModeCardHandler(cardMsg CardMsg,
	m MessageHandler) CardHandlerFunc {
	return func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {

		if cardMsg.Kind == AIModeChooseKind {
			newCard, err, done := CommonProcessAIMode(cardMsg, cardAction,
				m.sessionCache)
			if done {
				return newCard, err
			}
			return nil, nil
		}
		return nil, ErrNextHandler
	}
}

// CommonProcessAIMode is the common process for choosing AI mode
func CommonProcessAIMode(msg CardMsg, cardAction *larkcard.CardAction,
	cache services.SessionServiceCacheInterface) (interface{},
	error, bool) {
	option := cardAction.Action.Option
	cache.SetAIMode(msg.SessionId, openai.AIModeMap[option])

	// Return a confirmation card instead of trying to send a message
	newCard, _ := newSendCard(
		withHeader("Divergent Mode Selection", larkcard.TemplateIndigo),
		withMainMd("Selected divergent mode: **"+option+"**"),
		withNote("The AI mode has been updated. You can continue chatting."),
	)
	return newCard, nil, true
}

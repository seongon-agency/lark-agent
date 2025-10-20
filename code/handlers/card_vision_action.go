package handlers

import (
	"context"
	"fmt"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"start-feishubot/services"
)

func NewVisionResolutionHandler(cardMsg CardMsg,
	m MessageHandler) CardHandlerFunc {
	return func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		if cardMsg.Kind == VisionStyleKind {
			newCard, err, done := CommonProcessVisionStyle(cardMsg, cardAction, m.sessionCache)
			if done {
				return newCard, err
			}
			return nil, nil
		}
		return nil, ErrNextHandler
	}
}
func NewVisionModeChangeHandler(cardMsg CardMsg,
	m MessageHandler) CardHandlerFunc {
	return func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		if cardMsg.Kind == VisionModeChangeKind {
			newCard, err, done := CommonProcessVisionModeChange(cardMsg, m.sessionCache)
			if done {
				return newCard, err
			}
			return nil, nil
		}
		return nil, ErrNextHandler
	}
}

func CommonProcessVisionStyle(msg CardMsg,
	cardAction *larkcard.CardAction,
	cache services.SessionServiceCacheInterface) (interface{}, error, bool) {
	option := cardAction.Action.Option
	fmt.Println(larkcore.Prettify(msg))
	cache.SetVisionDetail(msg.SessionId, services.VisionDetail(option))

	// Return a confirmation card
	newCard, _ := newSendCard(
		withHeader("🕵️ Image Reasoning Mode", larkcard.TemplateBlue),
		withMainMd("Image resolution adjusted to: **"+option+"**"),
		withVisionDetailLevelBtn(&msg.SessionId),
		withNote("You can continue to adjust settings or upload images for analysis."),
	)
	return newCard, nil, true
}

func CommonProcessVisionModeChange(cardMsg CardMsg,
	session services.SessionServiceCacheInterface) (
	interface{}, error, bool) {
	if cardMsg.Value == "1" {

		sessionId := cardMsg.SessionId
		session.Clear(sessionId)
		session.SetMode(sessionId,
			services.ModeVision)
		session.SetVisionDetail(sessionId,
			services.VisionDetailLow)

		newCard, _ :=
			newSendCard(
				withHeader("🕵️️ Entered image reasoning mode", larkcard.TemplateBlue),
				withVisionDetailLevelBtn(&sessionId),
				withNote("Reminder: Reply with images to let LLM reason about the image content with you."))
		return newCard, nil, true
	}
	if cardMsg.Value == "0" {
		newCard, _ := newSendCard(
			withHeader("🎒 Bot Reminder", larkcard.TemplateGreen),
			withMainMd("Context information for this topic is still retained"),
			withNote("We can continue discussing this topic, looking forward to chatting with you. If you have other questions or topics you'd like to discuss, please let me know"),
		)
		return newCard, nil, true
	}
	return nil, nil, false
}

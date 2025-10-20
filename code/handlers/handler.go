package handlers

import (
	"context"
	"fmt"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"start-feishubot/logger"
	"strings"

	"start-feishubot/initialization"
	"start-feishubot/services"
	"start-feishubot/services/openai"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// Chain of responsibility pattern
func chain(data *ActionInfo, actions ...Action) bool {
	for _, v := range actions {
		if !v.Execute(data) {
			return false
		}
	}
	return true
}

type MessageHandler struct {
	sessionCache services.SessionServiceCacheInterface
	msgCache     services.MsgCacheInterface
	gpt          *openai.ChatGPT
	config       initialization.Config
}

func (m MessageHandler) cardHandler(ctx context.Context,
	cardAction *larkcard.CardAction) (interface{}, error) {
	messageHandler := NewCardHandler(m)
	return messageHandler(ctx, cardAction)
}

func judgeMsgType(event *larkim.P2MessageReceiveV1) (string, error) {
	msgType := event.Event.Message.MessageType

	switch *msgType {
	case "text", "image", "audio", "post":
		return *msgType, nil
	default:
		return "", fmt.Errorf("unknown message type: %v", *msgType)
	}
}

func (m MessageHandler) msgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	handlerType := judgeChatType(event)
	logger.Debug("handlerType", handlerType)
	if handlerType == "otherChat" {
		fmt.Println("unknown chat type")
		return nil
	}
	logger.Debug("Received message: ", larkcore.Prettify(event.Event.Message))

	msgType, err := judgeMsgType(event)
	if err != nil {
		fmt.Printf("error getting message type: %v\n", err)
		return nil
	}

	content := event.Event.Message.Content
	msgId := event.Event.Message.MessageId
	rootId := event.Event.Message.RootId
	chatId := event.Event.Message.ChatId
	mention := event.Event.Message.Mentions

	sessionId := rootId
	if sessionId == nil || *sessionId == "" {
		sessionId = msgId
	}
	msgInfo := MsgInfo{
		handlerType: handlerType,
		msgType:     msgType,
		msgId:       msgId,
		chatId:      chatId,
		qParsed:     strings.Trim(parseContent(*content, msgType), " "),
		fileKey:     parseFileKey(*content),
		imageKey:    parseImageKey(*content),
		imageKeys:   parsePostImageKeys(*content),
		sessionId:   sessionId,
		mention:     mention,
	}
	data := &ActionInfo{
		ctx:     &ctx,
		handler: &m,
		info:    &msgInfo,
	}
	actions := []Action{
		&ProcessedUniqueAction{}, //Avoid duplicate processing
		&ProcessMentionAction{},  //Check if bot should be invoked
		&AudioAction{},           //Audio processing
		&ClearAction{},           //Clear message processing
		&VisionAction{},          //Image reasoning processing
		&PicAction{},             //Picture processing
		&AIModeAction{},          //Mode switching processing
		&RoleListAction{},        //Role list processing
		&HelpAction{},            //Help processing
		&BalanceAction{},         //Balance processing
		&RolePlayAction{},        //Role play processing
		&MessageAction{},         //Message processing
		&EmptyAction{},           //Empty message processing
		&StreamMessageAction{},   //Stream message processing
	}
	chain(data, actions...)
	return nil
}

var _ MessageHandlerInterface = (*MessageHandler)(nil)

func NewMessageHandler(gpt *openai.ChatGPT,
	config initialization.Config) MessageHandlerInterface {
	return &MessageHandler{
		sessionCache: services.GetSessionCache(),
		msgCache:     services.GetMsgCache(),
		gpt:          gpt,
		config:       config,
	}
}

func (m MessageHandler) judgeIfMentionMe(mention []*larkim.
	MentionEvent) bool {
	if len(mention) != 1 {
		return false
	}
	return *mention[0].Name == m.config.FeishuBotName
}

func AzureModeCheck(a *ActionInfo) bool {
	if a.handler.config.AzureOn {
		//sendMsg(*a.ctx, "Azure Openai 接口下，暂不支持此功能", a.info.chatId)
		return false
	}
	return true
}

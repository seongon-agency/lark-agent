package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"start-feishubot/logger"

	"start-feishubot/initialization"
	"start-feishubot/services"
	"start-feishubot/services/openai"

	"github.com/google/uuid"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type CardKind string
type CardChatType string

var (
	ClearCardKind        = CardKind("clear")            // Clear context
	PicModeChangeKind    = CardKind("pic_mode_change")  // Switch image creation mode
	VisionModeChangeKind = CardKind("vision_mode")      // Switch image analysis mode
	PicResolutionKind    = CardKind("pic_resolution")   // Image resolution adjustment
	PicStyleKind         = CardKind("pic_style")        // Image style adjustment
	VisionStyleKind      = CardKind("vision_style")     // Image reasoning level adjustment
	PicTextMoreKind      = CardKind("pic_text_more")    // Regenerate image from text
	PicVarMoreKind       = CardKind("pic_var_more")     // Variant image
	RoleTagsChooseKind   = CardKind("role_tags_choose") // Built-in role tag selection
	RoleChooseKind       = CardKind("role_choose")      // Built-in role selection
	AIModeChooseKind     = CardKind("ai_mode_choose")   // AI mode selection
)

var (
	GroupChatType = CardChatType("group")
	UserChatType  = CardChatType("personal")
)

type CardMsg struct {
	Kind      CardKind
	ChatType  CardChatType
	Value     interface{}
	SessionId string
	MsgId     string
}

type MenuOption struct {
	value string
	label string
}

func replyCard(ctx context.Context,
	msgId *string,
	cardContent string,
) error {
	client := initialization.GetLarkClient()
	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeInteractive).
			Uuid(uuid.New().String()).
			Content(cardContent).
			Build()).
		Build())

	// Handle errors
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Server-side error handling
	if !resp.Success() {
		logger.Errorf("Server error resp code[%v], msg [%v] requestId [%v] ", resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func newSendCard(
	header *larkcard.MessageCardHeader,
	elements ...larkcard.MessageCardElement) (string,
	error) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(false).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// Card message body
	cardContent, err := larkcard.NewMessageCard().
		Config(config).
		Header(header).
		Elements(
			aElementPool,
		).
		String()
	return cardContent, err
}

func newSimpleSendCard(
	elements ...larkcard.MessageCardElement) (string,
	error) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(false).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// Card message body
	cardContent, err := larkcard.NewMessageCard().
		Config(config).
		Elements(
			aElementPool,
		).
		String()
	return cardContent, err
}

// withSplitLine used to generate divider lines
func withSplitLine() larkcard.MessageCardElement {
	splitLine := larkcard.NewMessageCardHr().
		Build()
	return splitLine
}

// withHeader used to generate message headers
func withHeader(title string, color string) *larkcard.
	MessageCardHeader {
	if title == "" {
		title = "🤖️ Bot Reminder"
	}
	header := larkcard.NewMessageCardHeader().
		Template(color).
		Title(larkcard.NewMessageCardPlainText().
			Content(title).
			Build()).
		Build()
	return header
}

// withNote used to generate plain text footnotes
func withNote(note string) larkcard.MessageCardElement {
	noteElement := larkcard.NewMessageCardNote().
		Elements([]larkcard.MessageCardNoteElement{larkcard.NewMessageCardPlainText().
			Content(note).
			Build()}).
		Build()
	return noteElement
}

// withMainMd used to generate markdown message body
func withMainMd(msg string) larkcard.MessageCardElement {
	msg, i := processMessage(msg)
	msg = processNewLine(msg)
	if i != nil {
		return nil
	}
	mainElement := larkcard.NewMessageCardDiv().
		Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardLarkMd().
				Content(msg).
				Build()).
			IsShort(true).
			Build()}).
		Build()
	return mainElement
}

// withMainText used to generate plain text message body
func withMainText(msg string) larkcard.MessageCardElement {
	msg, i := processMessage(msg)
	msg = cleanTextBlock(msg)
	if i != nil {
		return nil
	}
	mainElement := larkcard.NewMessageCardDiv().
		Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardPlainText().
				Content(msg).
				Build()).
			IsShort(false).
			Build()}).
		Build()
	return mainElement
}

func withImageDiv(imageKey string) larkcard.MessageCardElement {
	imageElement := larkcard.NewMessageCardImage().
		ImgKey(imageKey).
		Alt(larkcard.NewMessageCardPlainText().Content("").
			Build()).
		Preview(true).
		Mode(larkcard.MessageCardImageModelCropCenter).
		CompactWidth(true).
		Build()
	return imageElement
}

// withMdAndExtraBtn used to generate message body with extra buttons
func withMdAndExtraBtn(msg string, btn *larkcard.
	MessageCardEmbedButton) larkcard.MessageCardElement {
	msg, i := processMessage(msg)
	msg = processNewLine(msg)
	if i != nil {
		return nil
	}
	mainElement := larkcard.NewMessageCardDiv().
		Fields(
			[]*larkcard.MessageCardField{
				larkcard.NewMessageCardField().
					Text(larkcard.NewMessageCardLarkMd().
						Content(msg).
						Build()).
					IsShort(true).
					Build()}).
		Extra(btn).
		Build()
	return mainElement
}

func newBtn(content string, value map[string]interface{},
	typename larkcard.MessageCardButtonType) *larkcard.
	MessageCardEmbedButton {
	btn := larkcard.NewMessageCardEmbedButton().
		Type(typename).
		Value(value).
		Text(larkcard.NewMessageCardPlainText().
			Content(content).
			Build())
	return btn
}

func newMenu(
	placeHolder string,
	value map[string]interface{},
	options ...MenuOption,
) *larkcard.
	MessageCardEmbedSelectMenuStatic {
	var aOptionPool []*larkcard.MessageCardEmbedSelectOption
	for _, option := range options {
		aOption := larkcard.NewMessageCardEmbedSelectOption().
			Value(option.value).
			Text(larkcard.NewMessageCardPlainText().
				Content(option.label).
				Build())
		aOptionPool = append(aOptionPool, aOption)

	}
	btn := larkcard.NewMessageCardEmbedSelectMenuStatic().
		MessageCardEmbedSelectMenuStatic(larkcard.NewMessageCardEmbedSelectMenuBase().
			Options(aOptionPool).
			Placeholder(larkcard.NewMessageCardPlainText().
				Content(placeHolder).
				Build()).
			Value(value).
			Build()).
		Build()
	return btn
}

// Clear card buttons
func withClearDoubleCheckBtn(sessionID *string) larkcard.MessageCardElement {
	confirmBtn := newBtn("Confirm Clear", map[string]interface{}{
		"value":     "1",
		"kind":      ClearCardKind,
		"chatType":  UserChatType,
		"sessionId": *sessionID,
	}, larkcard.MessageCardButtonTypeDanger,
	)
	cancelBtn := newBtn("Let me think", map[string]interface{}{
		"value":     "0",
		"kind":      ClearCardKind,
		"sessionId": *sessionID,
		"chatType":  UserChatType,
	},
		larkcard.MessageCardButtonTypeDefault)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{confirmBtn, cancelBtn}).
		Layout(larkcard.MessageCardActionLayoutBisected.Ptr()).
		Build()

	return actions
}

func withPicModeDoubleCheckBtn(sessionID *string) larkcard.
	MessageCardElement {
	confirmBtn := newBtn("Switch Mode", map[string]interface{}{
		"value":     "1",
		"kind":      PicModeChangeKind,
		"chatType":  UserChatType,
		"sessionId": *sessionID,
	}, larkcard.MessageCardButtonTypeDanger,
	)
	cancelBtn := newBtn("Let me think", map[string]interface{}{
		"value":     "0",
		"kind":      PicModeChangeKind,
		"sessionId": *sessionID,
		"chatType":  UserChatType,
	},
		larkcard.MessageCardButtonTypeDefault)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{confirmBtn, cancelBtn}).
		Layout(larkcard.MessageCardActionLayoutBisected.Ptr()).
		Build()

	return actions
}
func withVisionModeDoubleCheckBtn(sessionID *string) larkcard.
	MessageCardElement {
	confirmBtn := newBtn("Switch Mode", map[string]interface{}{
		"value":     "1",
		"kind":      VisionModeChangeKind,
		"chatType":  UserChatType,
		"sessionId": *sessionID,
	}, larkcard.MessageCardButtonTypeDanger,
	)
	cancelBtn := newBtn("Let me think", map[string]interface{}{
		"value":     "0",
		"kind":      VisionModeChangeKind,
		"sessionId": *sessionID,
		"chatType":  UserChatType,
	},
		larkcard.MessageCardButtonTypeDefault)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{confirmBtn, cancelBtn}).
		Layout(larkcard.MessageCardActionLayoutBisected.Ptr()).
		Build()

	return actions
}

func withOneBtn(btn *larkcard.MessageCardEmbedButton) larkcard.
	MessageCardElement {
	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{btn}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
	return actions
}

// New conversation button

func withPicResolutionBtn(sessionID *string) larkcard.
	MessageCardElement {
	resolutionMenu := newMenu("Default Resolution",
		map[string]interface{}{
			"value":     "0",
			"kind":      PicResolutionKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		// dall-e-2 256, 512, 1024
		//MenuOption{
		//	label: "256x256",
		//	value: string(services.Resolution256),
		//},
		//MenuOption{
		//	label: "512x512",
		//	value: string(services.Resolution512),
		//},
		// dall-e-3
		MenuOption{
			label: "1024x1024",
			value: string(services.Resolution1024),
		},
		MenuOption{
			label: "1024x1792",
			value: string(services.Resolution10241792),
		},
		MenuOption{
			label: "1792x1024",
			value: string(services.Resolution17921024),
		},
	)

	styleMenu := newMenu("Style",
		map[string]interface{}{
			"value":     "0",
			"kind":      PicStyleKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		MenuOption{
			label: "Vivid Style",
			value: string(services.PicStyleVivid),
		},
		MenuOption{
			label: "Natural Style",
			value: string(services.PicStyleNatural),
		},
	)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{resolutionMenu, styleMenu}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
	return actions
}

func withVisionDetailLevelBtn(sessionID *string) larkcard.
	MessageCardElement {
	detailMenu := newMenu("Select image resolution, default is high",
		map[string]interface{}{
			"value":     "0",
			"kind":      VisionStyleKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		MenuOption{
			label: "High",
			value: string(services.VisionDetailHigh),
		},
		MenuOption{
			label: "Low",
			value: string(services.VisionDetailLow),
		},
	)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{detailMenu}).
		Layout(larkcard.MessageCardActionLayoutBisected.Ptr()).
		Build()

	return actions
}
func withRoleTagsBtn(sessionID *string, tags ...string) larkcard.
	MessageCardElement {
	var menuOptions []MenuOption

	for _, tag := range tags {
		menuOptions = append(menuOptions, MenuOption{
			label: tag,
			value: tag,
		})
	}
	cancelMenu := newMenu("Select Role Category",
		map[string]interface{}{
			"value":     "0",
			"kind":      RoleTagsChooseKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		menuOptions...,
	)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{cancelMenu}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
	return actions
}

func withRoleBtn(sessionID *string, titles ...string) larkcard.
	MessageCardElement {
	var menuOptions []MenuOption

	for _, tag := range titles {
		menuOptions = append(menuOptions, MenuOption{
			label: tag,
			value: tag,
		})
	}
	cancelMenu := newMenu("View Built-in Roles",
		map[string]interface{}{
			"value":     "0",
			"kind":      RoleChooseKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		menuOptions...,
	)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{cancelMenu}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
	return actions
}

func withAIModeBtn(sessionID *string, aiModeStrs []string) larkcard.MessageCardElement {
	var menuOptions []MenuOption
	for _, label := range aiModeStrs {
		menuOptions = append(menuOptions, MenuOption{
			label: label,
			value: label,
		})
	}

	cancelMenu := newMenu("Select Mode",
		map[string]interface{}{
			"value":     "0",
			"kind":      AIModeChooseKind,
			"sessionId": *sessionID,
			"msgId":     *sessionID,
		},
		menuOptions...,
	)

	actions := larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{cancelMenu}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
	return actions
}

func replyMsg(ctx context.Context, msg string, msgId *string) error {
	msg, i := processMessage(msg)
	if i != nil {
		return i
	}
	client := initialization.GetLarkClient()
	content := larkim.NewTextMsgBuilder().
		Text(msg).
		Build()

	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// Handle errors
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Server-side error handling
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func uploadImage(base64Str string) (*string, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	client := initialization.GetLarkClient()
	resp, err := client.Im.Image.Create(context.Background(),
		larkim.NewCreateImageReqBuilder().
			Body(larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(bytes.NewReader(imageBytes)).
				Build()).
			Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.ImageKey, nil
}

func replyImage(ctx context.Context, ImageKey *string,
	msgId *string) error {
	//fmt.Println("sendMsg", ImageKey, msgId)

	msgImage := larkim.MessageImage{ImageKey: *ImageKey}
	content, err := msgImage.String()
	if err != nil {
		fmt.Println(err)
		return err
	}
	client := initialization.GetLarkClient()

	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeImage).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// Handle errors
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Server-side error handling
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func replayImageCardByBase64(ctx context.Context, base64Str string,
	msgId *string, sessionId *string, question string) error {
	imageKey, err := uploadImage(base64Str)
	if err != nil {
		return err
	}
	//example := "img_v2_041b28e3-5680-48c2-9af2-497ace79333g"
	//imageKey := &example
	//fmt.Println("imageKey", *imageKey)
	err = sendImageCard(ctx, *imageKey, msgId, sessionId, question)
	if err != nil {
		return err
	}
	return nil
}

func replayImagePlainByBase64(ctx context.Context, base64Str string,
	msgId *string) error {
	imageKey, err := uploadImage(base64Str)
	if err != nil {
		return err
	}
	//example := "img_v2_041b28e3-5680-48c2-9af2-497ace79333g"
	//imageKey := &example
	//fmt.Println("imageKey", *imageKey)
	err = replyImage(ctx, imageKey, msgId)
	if err != nil {
		return err
	}
	return nil
}

func replayVariantImageByBase64(ctx context.Context, base64Str string,
	msgId *string, sessionId *string) error {
	imageKey, err := uploadImage(base64Str)
	if err != nil {
		return err
	}
	//example := "img_v2_041b28e3-5680-48c2-9af2-497ace79333g"
	//imageKey := &example
	//fmt.Println("imageKey", *imageKey)
	err = sendVarImageCard(ctx, *imageKey, msgId, sessionId)
	if err != nil {
		return err
	}
	return nil
}

func sendMsg(ctx context.Context, msg string, chatId *string) error {
	//fmt.Println("sendMsg", msg, chatId)
	msg, i := processMessage(msg)
	if i != nil {
		return i
	}
	client := initialization.GetLarkClient()
	content := larkim.NewTextMsgBuilder().
		Text(msg).
		Build()

	//fmt.Println("content", content)

	resp, err := client.Im.Message.Create(ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(*chatId).
			Content(content).
			Build()).
		Build())

	// Handle errors
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Server-side error handling
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func sendClearCacheCheckCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🆑 Bot Reminder", larkcard.TemplateBlue),
		withMainMd("Are you sure you want to clear the conversation context?"),
		withNote("Please note, this will start a brand new conversation and you won't be able to use historical information from previous topics"),
		withClearDoubleCheckBtn(sessionId))
	replyCard(ctx, msgId, newCard)
}

func sendSystemInstructionCard(ctx context.Context,
	sessionId *string, msgId *string, content string) {
	newCard, _ := newSendCard(
		withHeader("🥷  Entered Role-Playing Mode", larkcard.TemplateIndigo),
		withMainText(content),
		withNote("Please note, this will start a brand new conversation and you won't be able to use historical information from previous topics"))
	replyCard(ctx, msgId, newCard)
}

func sendPicCreateInstructionCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🖼️ Entered Image Creation Mode", larkcard.TemplateBlue),
		withPicResolutionBtn(sessionId),
		withNote("Reminder: Reply with text or images to let AI generate related pictures."))
	replyCard(ctx, msgId, newCard)
}

func sendVisionInstructionCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🕵️️ Entered Image Analysis Mode", larkcard.TemplateBlue),
		withVisionDetailLevelBtn(sessionId),
		withNote("Reminder: Reply with images to let the LLM analyze the image content with you."))
	replyCard(ctx, msgId, newCard)
}

func sendPicModeCheckCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🖼️ Bot Reminder", larkcard.TemplateBlue),
		withMainMd("Image received, enter image creation mode?"),
		withNote("Please note, this will start a brand new conversation and you won't be able to use historical information from previous topics"),
		withPicModeDoubleCheckBtn(sessionId))
	replyCard(ctx, msgId, newCard)
}
func sendVisionModeCheckCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🕵️ Bot Reminder", larkcard.TemplateBlue),
		withMainMd("Image detected, enter image analysis mode?"),
		withNote("Please note, this will start a brand new conversation and you won't be able to use historical information from previous topics"),
		withVisionModeDoubleCheckBtn(sessionId))
	replyCard(ctx, msgId, newCard)
}

func sendNewTopicCard(ctx context.Context,
	sessionId *string, msgId *string, content string) {
	newCard, _ := newSendCard(
		withHeader("👻️ Started New Topic", larkcard.TemplateBlue),
		withMainText(content),
		withNote("Reminder: Click the dialogue box to reply and maintain topic continuity"))
	replyCard(ctx, msgId, newCard)
}

func sendOldTopicCard(ctx context.Context,
	sessionId *string, msgId *string, content string) {
	newCard, _ := newSendCard(
		withHeader("🔃️ Contextual Topic", larkcard.TemplateBlue),
		withMainText(content),
		withNote("Reminder: Click the dialogue box to reply and maintain topic continuity"))
	replyCard(ctx, msgId, newCard)
}

func sendVisionTopicCard(ctx context.Context,
	sessionId *string, msgId *string, content string) {
	newCard, _ := newSendCard(
		withHeader("🕵️ Image Analysis Result", larkcard.TemplateBlue),
		withMainText(content),
		withNote("Let the LLM analyze the image content with you~"))
	replyCard(ctx, msgId, newCard)
}

func sendHelpCard(ctx context.Context,
	sessionId *string, msgId *string) {
	newCard, _ := newSendCard(
		withHeader("🎒 Need Help?", larkcard.TemplateBlue),
		withMainMd("**🤠 Hello! I'm an intelligent assistant based on OpenAI!**"),
		withSplitLine(),
		withMdAndExtraBtn(
			"** 🆑 Clear Topic Context**\nReply with *clear* or */clear*",
			newBtn("Clear Now", map[string]interface{}{
				"value":     "1",
				"kind":      ClearCardKind,
				"chatType":  UserChatType,
				"sessionId": *sessionId,
			}, larkcard.MessageCardButtonTypeDanger)),
		withSplitLine(),
		withMainMd("🤖 **Divergent Mode Selection**\nReply with *ai mode* or */ai_mode*"),
		withSplitLine(),
		withMainMd("🛖 **Built-in Role List**\nReply with *roles* or */roles*"),
		withSplitLine(),
		withMainMd("🥷 **Role-Playing Mode**\nReply with *role play* or */system* + space + role info"),
		withSplitLine(),
		withMainMd("🎤 **AI Voice Chat**\nDirectly send voice messages in private chat mode"),
		withSplitLine(),
		withMainMd("🎨 **Image Creation Mode**\nReply with *picture* or */picture*"),
		withSplitLine(),
		withMainMd("🕵️ **Image Analysis Mode**\nReply with *vision* or */vision*"),
		withSplitLine(),
		withMainMd("🎰 **Token Balance Query**\nReply with *balance* or */balance*"),
		withSplitLine(),
		withMainMd("🔃️ **History Topic Restore** 🚧\nEnter topic reply details page, reply with *restore* or */reload*"),
		withSplitLine(),
		withMainMd("📤 **Export Topic Content** 🚧\nReply with *export* or */export*"),
		withSplitLine(),
		withMainMd("🎰 **Continuous Dialogue & Multi-Topic Mode**\nClick the dialogue box to reply and maintain topic continuity. Meanwhile, ask separately to start a new topic"),
		withSplitLine(),
		withMainMd("🎒 **Need More Help?**\nReply with *help* or */help*"),
	)
	replyCard(ctx, msgId, newCard)
}

func sendImageCard(ctx context.Context, imageKey string,
	msgId *string, sessionId *string, question string) error {
	newCard, _ := newSimpleSendCard(
		withImageDiv(imageKey),
		withSplitLine(),
		// One more
		withOneBtn(newBtn("One More", map[string]interface{}{
			"value":     question,
			"kind":      PicTextMoreKind,
			"chatType":  UserChatType,
			"msgId":     *msgId,
			"sessionId": *sessionId,
		}, larkcard.MessageCardButtonTypePrimary)),
	)
	replyCard(ctx, msgId, newCard)
	return nil
}

func sendVarImageCard(ctx context.Context, imageKey string,
	msgId *string, sessionId *string) error {
	newCard, _ := newSimpleSendCard(
		withImageDiv(imageKey),
		withSplitLine(),
		// One more
		withOneBtn(newBtn("One More", map[string]interface{}{
			"value":     imageKey,
			"kind":      PicVarMoreKind,
			"chatType":  UserChatType,
			"msgId":     *msgId,
			"sessionId": *sessionId,
		}, larkcard.MessageCardButtonTypePrimary)),
	)
	replyCard(ctx, msgId, newCard)
	return nil
}

func sendBalanceCard(ctx context.Context, msgId *string,
	balance openai.BalanceResponse) {
	newCard, _ := newSendCard(
		withHeader("🎰️ Balance Query", larkcard.TemplateBlue),
		withMainMd(fmt.Sprintf("Total Quota: %.2f$", balance.TotalGranted)),
		withMainMd(fmt.Sprintf("Used Quota: %.2f$", balance.TotalUsed)),
		withMainMd(fmt.Sprintf("Available Quota: %.2f$",
			balance.TotalAvailable)),
		withNote(fmt.Sprintf("Validity Period: %s - %s",
			balance.EffectiveAt.Format("2006-01-02 15:04:05"),
			balance.ExpiresAt.Format("2006-01-02 15:04:05"))),
	)
	replyCard(ctx, msgId, newCard)
}

func SendRoleTagsCard(ctx context.Context,
	sessionId *string, msgId *string, roleTags []string) {
	newCard, _ := newSendCard(
		withHeader("🛖 Please Select Role Category", larkcard.TemplateIndigo),
		withRoleTagsBtn(sessionId, roleTags...),
		withNote("Reminder: Select the role category so we can recommend more related roles for you."))
	err := replyCard(ctx, msgId, newCard)
	if err != nil {
		logger.Errorf("Error selecting role %v", err)
	}
}

func SendRoleListCard(ctx context.Context,
	sessionId *string, msgId *string, roleTag string, roleList []string) {
	newCard, _ := newSendCard(
		withHeader("🛖 Role List"+" - "+roleTag, larkcard.TemplateIndigo),
		withRoleBtn(sessionId, roleList...),
		withNote("Reminder: Select a built-in scenario to quickly enter role-playing mode."))
	replyCard(ctx, msgId, newCard)
}

func SendAIModeListsCard(ctx context.Context,
	sessionId *string, msgId *string, aiModeStrs []string) {
	newCard, _ := newSendCard(
		withHeader("🤖 Divergent Mode Selection", larkcard.TemplateIndigo),
		withAIModeBtn(sessionId, aiModeStrs),
		withNote("Reminder: Select a built-in mode to help AI better understand your needs."))
	replyCard(ctx, msgId, newCard)
}

func sendOnProcessCard(ctx context.Context,
	sessionId *string, msgId *string, ifNewTopic bool) (*string,
	error) {
	var newCard string
	if ifNewTopic {
		newCard, _ = newSendCard(
			withHeader("👻️ Started New Topic", larkcard.TemplateBlue),
			withNote("Thinking, please wait..."))
	} else {
		newCard, _ = newSendCard(
			withHeader("🔃️ Contextual Topic", larkcard.TemplateBlue),
			withNote("Thinking, please wait..."))
	}

	id, err := replyCardWithBackId(ctx, msgId, newCard)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func updateTextCard(ctx context.Context, msg string,
	msgId *string, ifNewTopic bool) error {
	var newCard string
	if ifNewTopic {
		newCard, _ = newSendCard(
			withHeader("👻️ Started New Topic", larkcard.TemplateBlue),
			withMainText(msg),
			withNote("Generating, please wait..."))
	} else {
		newCard, _ = newSendCard(
			withHeader("🔃️ Contextual Topic", larkcard.TemplateBlue),
			withMainText(msg),
			withNote("Generating, please wait..."))
	}
	err := PatchCard(ctx, msgId, newCard)
	if err != nil {
		return err
	}
	return nil
}
func updateFinalCard(
	ctx context.Context,
	msg string,
	msgId *string,
	ifNewSession bool,
) error {
	var newCard string
	if ifNewSession {
		newCard, _ = newSendCard(
			withHeader("👻️ Started New Topic", larkcard.TemplateBlue),
			withMainText(msg),
			withNote("Completed, you can continue asking questions or choose other functions."))
	} else {
		newCard, _ = newSendCard(
			withHeader("🔃️ Contextual Topic", larkcard.TemplateBlue),

			withMainText(msg),
			withNote("Completed, you can continue asking questions or choose other functions."))
	}
	err := PatchCard(ctx, msgId, newCard)
	if err != nil {
		return err
	}
	return nil
}

func newSendCardWithOutHeader(
	elements ...larkcard.MessageCardElement) (string, error) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(true).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// Card message body
	cardContent, err := larkcard.NewMessageCard().
		Config(config).
		Elements(
			aElementPool,
		).
		String()
	return cardContent, err
}

func PatchCard(ctx context.Context, msgId *string,
	cardContent string) error {
	//fmt.Println("sendMsg", msg, chatId)
	client := initialization.GetLarkClient()
	//content := larkim.NewTextMsgBuilder().
	//	Text(msg).
	//	Build()

	//fmt.Println("content", content)

	resp, err := client.Im.Message.Patch(ctx, larkim.NewPatchMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewPatchMessageReqBodyBuilder().
			Content(cardContent).
			Build()).
		Build())

	// Handle errors
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Server-side error handling
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func replyCardWithBackId(ctx context.Context,
	msgId *string,
	cardContent string,
) (*string, error) {
	client := initialization.GetLarkClient()
	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeInteractive).
			Uuid(uuid.New().String()).
			Content(cardContent).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}

	//ctx = context.WithValue(ctx, "SendMsgId", *resp.Data.MessageId)
	//SendMsgId := ctx.Value("SendMsgId")
	//pp.Println(SendMsgId)
	return resp.Data.MessageId, nil
}

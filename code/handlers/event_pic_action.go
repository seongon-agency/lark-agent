package handlers

import (
	"context"
	"fmt"
	"os"
	"start-feishubot/logger"

	"start-feishubot/initialization"
	"start-feishubot/services"
	"start-feishubot/services/openai"
	"start-feishubot/utils"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type PicAction struct { /*Picture*/
}

func (*PicAction) Execute(a *ActionInfo) bool {
	check := AzureModeCheck(a)
	if !check {
		return true
	}
	// Enable picture creation mode
	if _, foundPic := utils.EitherTrimEqual(a.info.qParsed,
		"/picture", "Picture Creation"); foundPic {
		a.handler.sessionCache.Clear(*a.info.sessionId)
		a.handler.sessionCache.SetMode(*a.info.sessionId,
			services.ModePicCreate)
		a.handler.sessionCache.SetPicResolution(*a.info.sessionId,
			services.Resolution1024)
		sendPicCreateInstructionCard(*a.ctx, a.info.sessionId,
			a.info.msgId)
		return false
	}

	mode := a.handler.sessionCache.GetMode(*a.info.sessionId)
	//fmt.Println("mode: ", mode)
	logger.Debug("MODE:", mode)
	// Received an image, and not in picture creation mode, prompt whether to switch to picture creation mode
	if a.info.msgType == "image" && mode != services.ModePicCreate {
		sendPicModeCheckCard(*a.ctx, a.info.sessionId, a.info.msgId)
		return false
	}

	if a.info.msgType == "image" && mode == services.ModePicCreate {
		//Save image
		imageKey := a.info.imageKey
		//fmt.Printf("fileKey: %s \n", imageKey)
		msgId := a.info.msgId
		//fmt.Println("msgId: ", *msgId)
		req := larkim.NewGetMessageResourceReqBuilder().MessageId(
			*msgId).FileKey(imageKey).Type("image").Build()
		resp, err := initialization.GetLarkClient().Im.MessageResource.Get(context.Background(), req)
		//fmt.Println(resp, err)
		if err != nil {
			//fmt.Println(err)
			replyMsg(*a.ctx, fmt.Sprintf("🤖️: Image download failed, please try again later. Error message: %v", err),
				a.info.msgId)
			return false
		}

		f := fmt.Sprintf("%s.png", imageKey)
		resp.WriteFile(f)
		defer os.Remove(f)
		resolution := a.handler.sessionCache.GetPicResolution(*a.
			info.sessionId)

		openai.ConvertJpegToPNG(f)
		openai.ConvertToRGBA(f, f)

		//Image verification
		err = openai.VerifyPngs([]string{f})
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf("🤖️: Unable to parse image, please send original image and try again~"),
				a.info.msgId)
			return false
		}
		bs64, err := a.handler.gpt.GenerateOneImageVariation(f, resolution)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️: Image generation failed, please try again later. Error message: %v", err), a.info.msgId)
			return false
		}
		replayImagePlainByBase64(*a.ctx, bs64, a.info.msgId)
		return false

	}

	// Generate image
	if mode == services.ModePicCreate {
		resolution := a.handler.sessionCache.GetPicResolution(*a.
			info.sessionId)
		style := a.handler.sessionCache.GetPicStyle(*a.
			info.sessionId)
		bs64, err := a.handler.gpt.GenerateOneImage(a.info.qParsed,
			resolution, style)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️: Image generation failed, please try again later. Error message: %v", err), a.info.msgId)
			return false
		}
		replayImageCardByBase64(*a.ctx, bs64, a.info.msgId, a.info.sessionId,
			a.info.qParsed)
		return false
	}

	return true
}

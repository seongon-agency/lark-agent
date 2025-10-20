package handlers

import (
	"context"
	"fmt"
	"os"

	"start-feishubot/initialization"
	"start-feishubot/utils/audio"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type AudioAction struct { /*Audio*/
}

func (*AudioAction) Execute(a *ActionInfo) bool {
	check := AzureModeCheck(a)
	if !check {
		return true
	}

	// Only parse audio in private chats, ignore in other chats
	if a.info.handlerType != UserHandler {
		return true
	}

	//Check if it's audio
	if a.info.msgType == "audio" {
		fileKey := a.info.fileKey
		//fmt.Printf("fileKey: %s \n", fileKey)
		msgId := a.info.msgId
		//fmt.Println("msgId: ", *msgId)
		req := larkim.NewGetMessageResourceReqBuilder().MessageId(
			*msgId).FileKey(fileKey).Type("file").Build()
		resp, err := initialization.GetLarkClient().Im.MessageResource.Get(context.Background(), req)
		//fmt.Println(resp, err)
		if err != nil {
			fmt.Println(err)
			return true
		}
		f := fmt.Sprintf("%s.ogg", fileKey)
		resp.WriteFile(f)
		defer os.Remove(f)

		//fmt.Println("f: ", f)
		output := fmt.Sprintf("%s.mp3", fileKey)
		// Wait for conversion to complete
		audio.OggToWavByPath(f, output)
		defer os.Remove(output)
		//fmt.Println("output: ", output)

		text, err := a.handler.gpt.AudioToText(output)
		if err != nil {
			fmt.Println(err)

			sendMsg(*a.ctx, fmt.Sprintf("🤖️: Audio conversion failed, please try again later. Error message: %v", err), a.info.msgId)
			return false
		}

		replyMsg(*a.ctx, fmt.Sprintf("🤖️：%s", text), a.info.msgId)
		//fmt.Println("text: ", text)
		a.info.qParsed = text
		return true
	}

	return true

}

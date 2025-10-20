package openai

import (
	"context"
	"fmt"
	"testing"
	"time"

	"start-feishubot/initialization"
)

func TestCompletions(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	msgs := []Messages{
		{Role: "system", Content: "You are a professional translator responsible for Chinese-English translation."},
		{Role: "user", Content: "Translate this: The assistant messages help store prior responses. They can also be written by a developer to help give examples of desired behavior."},
	}
	gpt := NewChatGPT(*config)
	resp, err := gpt.Completions(msgs, Balance)
	if err != nil {
		t.Errorf("TestCompletions failed with error: %v", err)
	}
	fmt.Println(resp.Content, resp.Role)
}

func TestVisionOnePic(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	content := []ContentType{
		{Type: "text", Text: "What’s in this image?", ImageURL: nil},
		{Type: "image_url", ImageURL: &ImageURL{
			URL: "https://resource.liaobots." +
				"com/1849d492904448a0ac17f975f0b7ca8b.jpg",
			Detail: "high",
		}},
	}

	msgs := []VisionMessages{
		{Role: "assistant", Content: content},
	}
	gpt := NewChatGPT(*config)
	resp, err := gpt.GetVisionInfo(msgs)
	if err != nil {
		t.Errorf("TestCompletions failed with error: %v", err)
	}
	fmt.Println(resp.Content, resp.Role)
}

func TestGenerateOneImage(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	gpt := NewChatGPT(*config)
	prompt := "a red apple"
	size := "256x256"
	imageURL, err := gpt.GenerateOneImage(prompt, size, "")
	if err != nil {
		t.Errorf("TestGenerateOneImage failed with error: %v", err)
	}
	if imageURL == "" {
		t.Errorf("TestGenerateOneImage returned empty imageURL")
	}
}

func TestAudioToText(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	gpt := NewChatGPT(*config)
	audio := "./test_file/test.wav"
	text, err := gpt.AudioToText(audio)
	if err != nil {
		t.Errorf("TestAudioToText failed with error: %v", err)
	}
	fmt.Printf("TestAudioToText returned text: %s \n", text)
	if text == "" {
		t.Errorf("TestAudioToText returned empty text")
	}

}

func TestVariateOneImage(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	gpt := NewChatGPT(*config)
	image := "./test_file/img.png"
	size := "256x256"
	//compressionType, err := GetImageCompressionType(image)
	//if err != nil {
	//	return
	//}
	//fmt.Println("compressionType: ", compressionType)
	ConvertToRGBA(image, image)
	err := VerifyPngs([]string{image})
	if err != nil {
		t.Errorf("TestVariateOneImage failed with error: %v", err)
		return
	}

	imageBs64, err := gpt.GenerateOneImageVariation(image, size)
	if err != nil {
		t.Errorf("TestVariateOneImage failed with error: %v", err)
	}
	//fmt.Printf("TestVariateOneImage returned imageBs64: %s \n", imageBs64)
	if imageBs64 == "" {
		t.Errorf("TestVariateOneImage returned empty imageURL")
	}
}

func TestVariateOneImageWithJpg(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	gpt := NewChatGPT(*config)
	image := "./test_file/test.jpg"
	size := "256x256"
	compressionType, err := GetImageCompressionType(image)
	if err != nil {
		return
	}
	fmt.Println("compressionType: ", compressionType)
	//ConvertJPGtoPNG(image)
	ConvertToRGBA(image, image)
	err = VerifyPngs([]string{image})
	if err != nil {
		t.Errorf("TestVariateOneImage failed with error: %v", err)
		return
	}

	imageBs64, err := gpt.GenerateOneImageVariation(image, size)
	if err != nil {
		t.Errorf("TestVariateOneImage failed with error: %v", err)
	}
	fmt.Printf("TestVariateOneImage returned imageBs64: %s \n", imageBs64)
	if imageBs64 == "" {
		t.Errorf("TestVariateOneImage returned empty imageURL")
	}
}

// Balance API has been deprecated
func TestChatGPT_GetBalance(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")
	gpt := NewChatGPT(*config)
	balance, err := gpt.GetBalance()
	if err != nil {
		t.Errorf("TestChatGPT_GetBalance failed with error: %v", err)
	}
	fmt.Println("balance: ", balance)
}

func TestChatGPT_streamChat(t *testing.T) {
	// Initialize configuration
	config := initialization.LoadConfig("../../config.yaml")

	// Prepare test cases
	testCases := []struct {
		msg        []Messages
		wantOutput string
		wantErr    bool
	}{
		{
			msg: []Messages{
				{
					Role:    "system",
					Content: "From now on, you need to become a workplace language master. You need to respond to questions from your boss in a tactful way, or make requests to leadership.",
				},
				{
					Role:    "user",
					Content: "Boss, I would like to request a day off",
				},
			},
			wantOutput: "",
			wantErr:    false,
		},
	}

	// Execute test cases
	for _, tc := range testCases {
		// Prepare input and output
		responseStream := make(chan string)
		ctx := context.Background()
		c := NewChatGPT(*config)

		// Start a goroutine to simulate streaming chat
		go func() {
			err := c.StreamChat(ctx, tc.msg, Balance, responseStream)
			if err != nil {
				t.Errorf("streamChat() error = %v, wantErr %v", err, tc.wantErr)
			}
		}()

		// Wait for output and check if it meets expectations
		select {
		case gotOutput := <-responseStream:
			fmt.Printf("gotOutput: %v\n", gotOutput)

		case <-time.After(5 * time.Second):
			t.Errorf("streamChat() timeout, expected output not received")
		}
	}
}

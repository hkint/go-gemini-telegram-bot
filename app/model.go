package app

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"go-gemini-telegram-bot/config"
)

const (
	TEXT_MODEL   = "gemini-pro"
	VISION_MODEL = "gemini-pro-vision"
)

var (
	client      *genai.Client
	textModel   *genai.GenerativeModel
	visionModel *genai.GenerativeModel
)

func InitModels() *genai.Client {
	initClient()
	log.Println("Initialized client")

	SafetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}

	textModel = client.GenerativeModel(TEXT_MODEL)
	textModel.SafetySettings = SafetySettings

	visionModel = client.GenerativeModel(VISION_MODEL)
	visionModel.SafetySettings = SafetySettings

	log.Println("Initialized models")
	return client
}

func initClient() {
	if client != nil {
		return
	}
	ctx := context.Background()
	var err error
	client, err = genai.NewClient(ctx, option.WithAPIKey(config.Env.Gemini_API_KEY))
	if err != nil {
		log.Fatal(err)
	}
}

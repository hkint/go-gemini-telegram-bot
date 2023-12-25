package app

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"go-gemini-telegram-bot/config"
)

var (
	client      *genai.Client
	textModel   *genai.GenerativeModel
	visionModel *genai.GenerativeModel
)

func init() {
	initClient()
	log.Println("Initialized client")

	SafetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}

	textModel = client.GenerativeModel("gemini-pro")
	textModel.SafetySettings = SafetySettings

	visionModel = client.GenerativeModel("gemini-pro-vision")
	visionModel.SafetySettings = SafetySettings

	log.Println("Initialized models")
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

func CloseClient() {
	if client != nil {
		client.Close()
	}
}

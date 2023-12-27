package app

import (
	"context"
	"errors"
	"fmt"
	"go-gemini-telegram-bot/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

const (
	TextModel   = "gemini-pro"
	VisionModel = "gemini-pro-vision"
)

var (
	ctx context.Context

	client *genai.Client

	textModel   *genai.GenerativeModel
	visionModel *genai.GenerativeModel

	modelMap = make(map[string]*genai.GenerativeModel, 2)

	chatSessionMap sync.Map
)

func InitModels() *genai.Client {
	if client != nil {
		return client
	}
	ctx = context.Background()
	var err error
	client, err = genai.NewClient(ctx, option.WithAPIKey(config.Env.GeminiApiKey))
	if err != nil {
		log.Fatal(err)
	}
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

	textModel = client.GenerativeModel(TextModel)
	visionModel = client.GenerativeModel(VisionModel)

	textModel.SafetySettings = SafetySettings
	visionModel.SafetySettings = SafetySettings

	modelMap[TextModel] = textModel
	modelMap[VisionModel] = visionModel

	log.Printf("Initialized models: %+v\n", modelMap)
	return client
}

func getChatSession(chatSessionID string) *genai.ChatSession {
	if val, ok := chatSessionMap.Load(chatSessionID); ok {
		return val.(*genai.ChatSession)
	}
	return nil
}

func setChatSession(chatSessionID string, chatSession *genai.ChatSession) {
	chatSessionMap.Store(chatSessionID, chatSession)
}

func handleChatSession(model *genai.GenerativeModel, chatSessionID string) (cs *genai.ChatSession) {
	if session := getChatSession(chatSessionID); session == nil {
		log.Printf("No chat session found, creating new one\n")
		cs = model.StartChat()
		setChatSession(chatSessionID, cs)
	} else {
		log.Printf("Chat session found, continue using it\n")
		cs = session
	}
	return
}

func generateSessionID(chatID int64, modelName string) string {
	return fmt.Sprintf("%d-%s", chatID, modelName)
}

func clearChatSession(sessionID string) bool {
	if session := getChatSession(sessionID); session != nil {
		session.History = nil
		return true
	} else {
		return false
	}
}

func getModelResponse(chatID int64, modelName string, parts []genai.Part) (response string) {
	sessionID := generateSessionID(chatID, modelName)
	cs := handleChatSession(modelMap[modelName], sessionID)

	iter := cs.SendMessageStream(ctx, parts...)

	handleResponse(iter, &response)
	return
}

func handleResponse(iter *genai.GenerateContentResponseIterator, response *string) {
	var buffer []string
	for {
		resp, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Println(err.Error())
			break
		}
		if resp != nil && len(resp.Candidates) > 0 {
			firstCandidate := resp.Candidates[0]
			if firstCandidate.Content != nil && len(firstCandidate.Content.Parts) > 0 {
				part := fmt.Sprint(firstCandidate.Content.Parts[0])
				buffer = append(buffer, part)
			} else {
				buffer = append(buffer, firstCandidate.FinishReason.String())
			}
		} else {
			buffer = append(buffer, "no content")
		}
	}
	*response = strings.Join(buffer, "")
}

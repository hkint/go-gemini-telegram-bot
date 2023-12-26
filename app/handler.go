package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

var ChatSessionMap sync.Map

func getChatSession(chatSessionID string) *genai.ChatSession {
	if val, ok := ChatSessionMap.Load(chatSessionID); ok {
		return val.(*genai.ChatSession)
	}
	return nil
}

func setChatSession(chatSessionID string, chatSession *genai.ChatSession) {
	ChatSessionMap.Store(chatSessionID, chatSession)
}

func handleChatSession(model *genai.GenerativeModel, chatSessionID string) (cs *genai.ChatSession) {
	if session := getChatSession(chatSessionID); session == nil {
		log.Printf("No chat session found, creating new one\n")
		cs = model.StartChat()
		setChatSession(chatSessionID, cs)
	} else {
		log.Printf("Chat session found, continue using it\n")
		cs = getChatSession(chatSessionID)
	}
	return
}

func generateSessionID(chatID int64, modelName string) string {
	return fmt.Sprintf("%d-%s", chatID, modelName)
}

func handleDefaultCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Try /start to get tips")
	bot.Send(msg)
}

func handleStartCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	user := update.Message.From
	startText := "Hi! " + user.FirstName + `
		Commands:
			/start - Show this message
			/new - Start a new chat session (clear previous contents)
		---
		Just send text or image to get response
	`
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, startText)
	bot.Send(msg)
}

func handleNewCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID
	textSessionID := generateSessionID(chatID, TEXT_MODEL)
	visionSessionID := generateSessionID(chatID, VISION_MODEL)

	info := "no chat session found, just send text or image"
	textClear, visionClear := clearChatSession(textSessionID), clearChatSession(visionSessionID)

	if textClear || visionClear {
		info = `Chat session cleared.`
	}

	msg := tgbotapi.NewMessage(chatID, info)
	bot.Send(msg)
}

func clearChatSession(sessionID string) bool {
	if session := getChatSession(sessionID); session != nil {
		setChatSession(sessionID, nil)
		return true
	} else {
		return false
	}
}

func handleTextMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	// Send "Generating..." message to user.
	msg := tgbotapi.NewMessage(chatID, "Waiting...")
	msg.ReplyToMessageID = update.Message.MessageID
	initMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return
	}

	// Simulate typing action.
	bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))

	// Generate a response by Gemini
	sessionID := generateSessionID(chatID, TEXT_MODEL)
	cs := handleChatSession(textModel, sessionID)
	ctx := context.Background()
	iter := cs.SendMessageStream(ctx, genai.Text(update.Message.Text))
	response := "hello"
	handleResponse(iter, &response)

	// Send the response back to the user.
	sendMessage(chatID, initMsg.MessageID, response, bot)

	time.Sleep(100 * time.Millisecond)
}

func handlePhotoMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	// Send "Generating..." message to user.
	msg := tgbotapi.NewMessage(chatID, "Waiting...")
	msg.ReplyToMessageID = update.Message.MessageID
	initMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return
	}

	// Simulate typing action.
	bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))

	photoMap := make(map[string]tgbotapi.PhotoSize)
	for _, photo := range update.Message.Photo {
		id := photo.FileID[:len(photo.FileID)-7]
		if existingPhoto, exists := photoMap[id]; !exists || photo.FileSize > existingPhoto.FileSize {
			photoMap[id] = photo
		}
	}

	images := make([]tgbotapi.PhotoSize, 0, len(photoMap))
	for _, photo := range photoMap {
		images = append(images, photo)
	}

	prompts := []genai.Part{}

	for _, img := range images {
		imgURL, err := bot.GetFileDirectURL(img.FileID)
		if err != nil {
			log.Printf("Error getting img URL: %v\n", err)
			return
		}
		imgRes, err := http.Get(imgURL)
		if err != nil {
			log.Printf("Error getting image response: %v\n", err)
			return
		}
		defer imgRes.Body.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, imgRes.Body)
		if err != nil {
			log.Fatal(err)
		}
		imgData := buf.Bytes()

		mimeType := http.DetectContentType(imgData)
		imageType := "jpeg"
		if strings.HasPrefix(mimeType, "image/") {
			imageType = strings.Split(mimeType, "/")[1]
		}
		prompts = append(prompts, genai.ImageData(imageType, imgData))
	}

	text := update.Message.Caption
	if text == "" {
		text = "Analyse these images and tell me what you see"
	}
	prompts = append(prompts, genai.Text(text))

	// Generate a response by Gemini
	sessionID := generateSessionID(chatID, VISION_MODEL)
	cs := handleChatSession(visionModel, sessionID)
	ctx := context.Background()
	iter := cs.SendMessageStream(ctx, prompts...)
	response := "hello"
	handleResponse(iter, &response)

	// Send the response back to the user.
	sendMessage(chatID, initMsg.MessageID, response, bot)

	time.Sleep(100 * time.Millisecond)
}

func handleResponse(iter *genai.GenerateContentResponseIterator, response *string) {
	buffer := ""
	for {
		res, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error getting response: %v\n", err)
			break
		}
		iterText := mergeResponse(res)

		buffer += iterText
	}
	*response = buffer
}

func mergeResponse(res *genai.GenerateContentResponse) string {
	var response string
	for _, cand := range res.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				text, ok := part.(genai.Text)
				if !ok {
					log.Printf("Error casting part to Text")
					continue
				}
				response += string(text)
			}
		}
	}
	return response
}

func sendMessage(chatID int64, initMsgID int, response string, bot *tgbotapi.BotAPI) {
	resp := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, response)
	edit := tgbotapi.NewEditMessageText(chatID, initMsgID, resp)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	edit.DisableWebPagePreview = true
	_, sendErr := bot.Send(edit)
	if sendErr != nil {
		log.Printf("Error sending message in ModeMarkdownV2: %v\n", sendErr)
		resp = tgbotapi.EscapeText(tgbotapi.ModeHTML, response)
		edit = tgbotapi.NewEditMessageText(chatID, initMsgID, resp)
		edit.ParseMode = tgbotapi.ModeHTML
		edit.DisableWebPagePreview = true
		_, sendErr = bot.Send(edit)
		if sendErr != nil {
			log.Printf("Error sending message in ModeHTML: %v\n", sendErr)
			edit.ParseMode = ""
			_, sendErr = bot.Send(edit)
			if sendErr != nil {
				log.Printf("Error sending message in ModeDefault: %v\n", sendErr)
			}
		}
	}
}

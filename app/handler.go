package app

import (
	"context"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

const telegramMessageLimit = 4096

var ChatSessionMap sync.Map

func getChatSession(chatID int64) *genai.ChatSession {
	if val, ok := ChatSessionMap.Load(chatID); ok {
		return val.(*genai.ChatSession)
	}
	return nil
}

func setChatSession(chatID int64, chatSession *genai.ChatSession) {
	ChatSessionMap.Store(chatID, chatSession)
}

func startNewChatSession(model *genai.GenerativeModel, chatID int64) *genai.ChatSession {
	chatSession := model.StartChat()
	setChatSession(chatID, chatSession)
	return chatSession
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

	var text string
	if session := getChatSession(chatID); session != nil {
		setChatSession(chatID, nil)
		text = `Chat session cleared.`
	} else {
		text = `New chat session started.`
	}

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func handleTextMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	var cs *genai.ChatSession
	if session := getChatSession(chatID); session == nil {
		log.Printf("No chat session found, creating new one\n")
		cs = startNewChatSession(textModel, chatID)
	} else {
		log.Printf("Chat session found, using it\n")
		cs = getChatSession(chatID)
	}

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

	text := update.Message.Text
	response := "..."

	// Generate a response by Gemini
	ctx := context.Background()
	iter := cs.SendMessageStream(ctx, genai.Text(text))
	handleResponse(iter, &response)
	tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, response)

	// Send the response back to the user.
	edit := tgbotapi.NewEditMessageText(chatID, initMsg.MessageID, response)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	_, err = bot.Send(edit)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return
	}

	time.Sleep(100 * time.Millisecond)
}

func handlePhotoMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	// Implement your image message handling logic here
}

func handleResponse(iter *genai.GenerateContentResponseIterator, response *string) {
	buffer := ""
	for {
		res, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		iterText := mergeResponse(res)
		if len(buffer)+len(iterText) > telegramMessageLimit {
			// over the limit, send the buffer and start a new one
			*response = buffer
			buffer = iterText
		} else {
			buffer += iterText
		}
	}
	if buffer != "" {
		*response = buffer
	}
}

func mergeResponse(resp *genai.GenerateContentResponse) string {
	var response string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				text, ok := part.(genai.Text)
				if !ok {
					log.Printf("Error casting part to Text")
				}
				response += string(text)
			}
		}
	}
	return response
}

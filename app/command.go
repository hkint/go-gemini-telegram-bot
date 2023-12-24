package app

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var ChatSessionMap sync.Map

func getSessionState(chatID int64) string {
	if val, ok := ChatSessionMap.Load(chatID); ok {
		return val.(string)
	}
	return ""
}

func setSessionState(chatID int64, state string) {
	ChatSessionMap.Store(chatID, state)
}

func DefaultCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Try /start to get tips")
	bot.Send(msg)
}

func StartCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	user := update.Message.From
	startText := "Hi! " + user.FirstName + `
		Commands:
		/start - Show this message
		/new - Start a new chat session (clear previous contents)
		---
		Just send text or image to get response`
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, startText)
	bot.Send(msg)
}

func NewChatCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	ChatSessionMap.Store(chatID, "new session started")

	msg := tgbotapi.NewMessage(chatID, "New chat session started.")
	bot.Send(msg)
}

func HandleText(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I received your message")
	bot.Send(msg)
}

func HandlePhoto(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	// Implement your image message handling logic here
}

package app

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func handleDefaultCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Try /start to get tips")
	sendMessage(bot, msg)
}

func handleStartCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	user := update.Message.From
	startText := "Hi! " + user.FirstName + ", Welcome to Gemini Bot! Send some text or image to start the conversation."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, startText)
	sendMessage(bot, msg)
}

func handleClearCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID
	textSessionID := generateSessionID(chatID, TextModel)
	visionSessionID := generateSessionID(chatID, VisionModel)

	info := "no chat session found, just send text or image"
	textClear, visionClear := clearChatSession(textSessionID), clearChatSession(visionSessionID)

	if textClear || visionClear {
		info = `Chat session cleared.`
	}

	msg := tgbotapi.NewMessage(chatID, info)
	sendMessage(bot, msg)
}

func handleHelpCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	helpInfo := `Commands: 
    /clear - Clear previous contents
    /help - Get help info
Just send text or image to get response`
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpInfo)
	sendMessage(bot, msg)
}

func handleTextMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	initMsgID, errFlag := instantReply(update, bot, chatID)
	if errFlag {
		return
	}

	generateResponse(bot, chatID, initMsgID, TextModel, genai.Text(update.Message.Text))
}

func handlePhotoMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID

	initMsgID, errFlag := instantReply(update, bot, chatID)
	if errFlag {
		return
	}

	var prompts []genai.Part
	errFlag = handlePhotoPrompts(update, bot, &prompts)
	if errFlag {
		return
	}

	generateResponse(bot, chatID, initMsgID, VisionModel, prompts...)
}

func instantReply(update tgbotapi.Update, bot *tgbotapi.BotAPI, chatID int64) (int, bool) {
	msg := tgbotapi.NewMessage(chatID, "Waiting...")
	msg.ReplyToMessageID = update.Message.MessageID
	initMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return 0, true
	}
	// Simulate typing action.
	_, _ = bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))

	return initMsg.MessageID, false
}

func handlePhotoPrompts(update tgbotapi.Update, bot *tgbotapi.BotAPI, prompts *[]genai.Part) bool {
	photo := update.Message.Photo[len(update.Message.Photo)-1]

	photoURL, err := getURL(bot, photo.FileID)
	if err != nil {
		return true
	}
	imgData, err := getImageData(photoURL)
	if err != nil {
		return true
	}
	imgType := getImageType(imgData)
	*prompts = append(*prompts, genai.ImageData(imgType, imgData))

	textPrompts := update.Message.Caption
	if textPrompts == "" {
		textPrompts = "Analyse these images and tell me what you see in Chinese"
	}
	*prompts = append(*prompts, genai.Text(textPrompts))
	return false
}

func getURL(bot *tgbotapi.BotAPI, fileID string) (string, error) {
	url, err := bot.GetFileDirectURL(fileID)
	if err != nil {
		log.Printf("Error getting img URL: %v\n", err)
		return "", err
	}
	return url, nil
}

func getImageData(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting image response: %v\n", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing image response: %v\n", err)
		}
	}(res.Body)

	imgData, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading image data: %v", err)
		return nil, err
	}

	return imgData, nil
}

func getImageType(data []byte) string {
	mimeType := http.DetectContentType(data)
	imageType := "jpeg"
	if strings.HasPrefix(mimeType, "image/") {
		imageType = strings.Split(mimeType, "/")[1]
	}
	return imageType
}

func generateResponse(bot *tgbotapi.BotAPI, chatID int64, initMsgID int, modelName string, parts ...genai.Part) {
	response := getModelResponse(chatID, modelName, parts)

	// Send the response back to the user.
	sendMessageInMarkdownV2(chatID, initMsgID, response, bot)

	time.Sleep(200 * time.Millisecond)
}

func sendMessage(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return
	}
}

func sendMessageInMarkdownV2(chatID int64, initMsgID int, response string, bot *tgbotapi.BotAPI) {
	edit := tgbotapi.NewEditMessageText(chatID, initMsgID, response)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	edit.DisableWebPagePreview = true
	_, sendErr := bot.Send(edit)
	if sendErr != nil {
		log.Printf("Error sending message in ModeMarkdownV2: %v\n", sendErr)
		edit.Text = tgbotapi.EscapeText(tgbotapi.ModeHTML, response)
		edit.ParseMode = tgbotapi.ModeHTML
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

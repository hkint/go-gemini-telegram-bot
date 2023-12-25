package app

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"go-gemini-telegram-bot/config"
)

func Start_bot() {
	bot, err := tgbotapi.NewBotAPI(config.Env.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					handleStartCommand(update, bot)
				case "new":
					handleNewCommand(update, bot)
				default:
					handleDefaultCommand(update, bot)
				}
			} else if update.Message.Text != "" {
				handleTextMessage(update, bot)
			} else if update.Message.Photo != nil {
				handlePhotoMessage(update, bot)
			}

		}

	}
}

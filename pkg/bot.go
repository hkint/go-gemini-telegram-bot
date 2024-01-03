package pkg

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"go-gemini-telegram-bot/config"
)

const (
	StartCommand = "start"
	ClearCommand = "clear"
	HelpCommand  = "help"
)

func StartBot() {
	bot, err := tgbotapi.NewBotAPI(config.Env.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.Env.DebugFlag

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set commands
	_, _ = bot.Request(tgbotapi.NewSetMyCommands([]tgbotapi.BotCommand{
		{
			Command:     ClearCommand,
			Description: "Clear previous contents and start a new chat",
		},
		{
			Command:     HelpCommand,
			Description: "Get help info",
		},
	}...))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Ignore any non-Message Updates
		if update.Message == nil {
			continue
		}
		// If set AllowedUsers, check if the user is allowed
		if len(config.Env.AllowedUsers) > 0 {
			if !contains(config.Env.AllowedUsers, update.Message.From.UserName) {
				log.Printf("User [ %s ] is not allowed to use this bot", update.Message.From.UserName)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not allowed to use this bot")
				_, _ = bot.Send(msg)
				continue
			}
		}
		// handle message
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case StartCommand:
				handleStartCommand(update, bot)
			case ClearCommand:
				handleClearCommand(update, bot)
			case HelpCommand:
				handleHelpCommand(update, bot)
			default:
				handleDefaultCommand(update, bot)
			}
		} else if update.Message.Photo != nil {
			handlePhotoMessage(update, bot)
		} else if update.Message.Text != "" {
			handleTextMessage(update, bot)
		}

	}
}

func contains(allowedUsers []string, userName string) bool {
	for _, allowedUser := range allowedUsers {
		if allowedUser == userName {
			return true
		}
	}
	return false
}

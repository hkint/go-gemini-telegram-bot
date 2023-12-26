package main

import (
	"go-gemini-telegram-bot/app"
)

func main() {
	client := app.InitModels()
	defer client.Close()
	app.Start_bot()
}

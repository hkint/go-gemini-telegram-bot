package main

import (
	"github.com/google/generative-ai-go/genai"
	"go-gemini-telegram-bot/bot"
	"log"
)

func main() {
	client := app.InitModels()
	defer func(client *genai.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)
	app.StartBot()
}

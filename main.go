package main

import (
	"github.com/google/generative-ai-go/genai"
	"go-gemini-telegram-bot/pkg"
	"log"
)

func main() {
	client := pkg.InitModels()
	defer func(client *genai.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)
	pkg.StartBot()
}

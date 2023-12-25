package config

import (
	"os"

	"log"

	"github.com/joho/godotenv"
)

var Env Environment

type Environment struct {
	BotToken       string
	Gemini_API_KEY string
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, trying to load from environment")
	}

	Env = Environment{
		BotToken:       getEnv("BOT_TOKEN", ""),
		Gemini_API_KEY: getEnv("Gemini_API_KEY", ""),
	}

	if Env.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set in environment variables or .env file")
	}
	if Env.Gemini_API_KEY == "" {
		log.Fatal("GEMINI_API_KEY must be set in environment variables or .env file")
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	loadEnv()
	log.Println("Loaded env")
}

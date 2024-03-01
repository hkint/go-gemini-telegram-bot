# go-gemini-telegram-bot

**A [Golang](https://golang.org/dl/) Telegram bot powered by Google's free `Gemini` LLM API**

## Overview

This Telegram bot uses Google's Gemini API to provide an AI assistant experience via Telegram. It is written in Golang.

## Features

- Chat with an AI assistant (powered by `Google Gemini`) on Telegram
- Support both `gemini-pro` and `gemini-pro-vision` models  
  (The `gemini-pro-vision` model is not optimized for multi-turn conversations)
- Restrict bot access to authorized users


## Configuration

### Requirements

- Google Gemini API key  
  Use your Google account to [create your API key](https://makersuite.google.com/app/apikey).
- Telegram bot token  
  Create a bot from Telegram [@BotFather](https://t.me/BotFather) and obtain an access token.

The bot is configured via environment variables or a `.env` file:

```
BOT_TOKEN = your_telegram_bot_token  
GEMINI_API_KEY = your_google_gemini_key
ALLOWED_USERS = username1,username2 # Optional, restrict bot access
```

See [.env.example](.env.example) for an example. Just copy or rename it to `.env`


## Building and Running

### Docker
- Pre-built images are available on GitHub Container Registry:
   ```
   docker pull ghcr.io/hkint/go-gemini-telegram-bot:latest
   ```
- Use Docker-Compose
   ```shell
   docker-compose up -d
   ```
  See [compose.yaml](compose.yaml) for details.

### Native

1. Install Go dependencies (Go version: 1.20 or higher)

   ```shell
   go mod tidy
   ```

2. Set environment variables (or use `.env` file, you can copy it from `.env_example`)
    ```shell
    export BOT_TOKEN='your_telegram_bot_token'
    export GEMINI_API_KEY='your_google_gemini_key'
    ```
3. Run the bot

   ```shell
   go run main.go
   ```


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

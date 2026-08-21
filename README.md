# ne Rugaysya Bot

a simple Telegram bot that detects bad words and sends a warning message.

## Features

- Detects bad words in incoming Telegram messages
- Replies with a warning: `Не ругайся!`
- Supports local execution and Docker deployment

## Requirements

- Go 1.26+
- Docker (optional)
- Telegram bot token

## Configuration

Create a `.env` file in the project root with the settings:

```env
BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
DB_PATH="/data/sqlite.db"
```

The bot also loads the bad word list from `badWords.txt`. Empty lines and lines starting with `#` are ignored.

## Local Run

```bash
go mod download
go run .
```

## Docker Compose

Use the existing `docker-compose.yml`:

```bash
docker compose up --build -d
```

## Notes

- The bot replies only once per message on first detected bad word.

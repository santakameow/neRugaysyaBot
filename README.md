# ne Rugaysya Bot

a simple Telegram bot that detects bad words and sends a warning message.

## Features

- Reads a list of forbidden words from `badWords.txt`
- Detects bad words in incoming Telegram messages
- Replies with a warning: `Не ругайся!`
- Supports local execution and Docker deployment

## Requirements

- Go 1.26+
- Docker (optional)
- Telegram bot token

## Configuration

Create a `.env` file in the project root with the bot token:

```env
BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
```

The bot also loads the bad word list from `badWords.txt`. Empty lines and lines starting with `#` are ignored.

## Local Run

```bash
go mod download
BOT_TOKEN=your_token_here go run main.go
```

Or with `.env` support:

```bash
go run main.go
```

## Docker

Build the image:

```bash
docker build -t nerugaysyabot:latest .
```

Run the container:

```bash
docker run --rm --env-file .env nerugaysyabot:latest
```

## Docker Compose

Use the existing `docker-compose.yml`:

```bash
docker compose up --build
```

## badWords.txt Format

- One word per line
- Case-insensitive matching
- Lines starting with `#` are treated as comments
- Example:

```text
# blocked words
badword
anotherbadword
```

## Notes

- The bot replies only once per message on first detected bad word.
- Ensure the `.env` file is not committed if it contains a real bot token.

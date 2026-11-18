# ne Rugaysya Bot

a simple Telegram bot that detects bad words and sends a warning message.

## Features

- Detects bad words in incoming Telegram messages
- Replies with a warning: `Не ругайся!`
- Supports local execution and Docker deployment

## Requirements

- Go 1.27.3+
- Docker (optional)
- Telegram bot token

## Configuration

Create a `.env` file in the project root with the settings:

```env
BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
DB_PATH="/data/sqlite.db"
OWNER_USER_ID=123456789
```

Set `OWNER_USER_ID` to **your own numeric Telegram user ID** (not the bot ID).
Only this user can run `/resetcount <user_id>` to set the specified user’s
swear count to zero, for example `/resetcount 987654321`. Other users are
silently ignored. Missing, invalid, or non-positive `OWNER_USER_ID` disables
resets. Restart the bot after changing configuration.

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

## Memory and backlog handling

Updates are handled by four workers with a 20-second deadline per update.
Long polling fetches at most 20 updates at a time and buffers four, providing
backpressure when catching up after downtime. SQLite uses one connection.

Telego uses Go's standard JSON implementation to avoid a startup allocation
spike in its default encoder when many updates are cloned concurrently.

Compose limits the container to 256 MiB and sets `GOMEMLIMIT=192MiB` (a soft
Go runtime memory target, not a hard cap). For local runs, set this variable
explicitly if desired. The container still restarts automatically after a crash.

## Notes

- The bot replies only once per message on first detected bad word.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	// load env
    err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	// context and correct exit
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	// create new bot
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	// get updates channel
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		log.Fatalf("failed to get updates: %v", err)
	}

	// bot handler
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Fatalf("failed to create bot handler: %v", err)
	}

	// get bad words list from badWords.txt
	badWords, err := os.ReadFile("badWords.txt")
	if err != nil {
		log.Fatalf("failed to read badWords.txt: %v", err)
	}

	var patterns []string
	for _, line := range strings.Split(string(badWords), "\n") {
		line = strings.TrimSpace(line)

		// skip empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	if len(patterns) == 0 {
		log.Fatal("no bad words patterns found in badWords.txt")
	}

	log.Printf("loaded %d bad words patterns", len(patterns))

	badWordsRegex, err := regexp.Compile("(?i)" + strings.Join(patterns, "|"))
	if err != nil {
		log.Fatalf("failed to compile bad words regex: %v", err)
	}

	// main handler
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// validate message exists
		if update.Message == nil {
			return nil
		}

		// check for bad words
		if badWordsRegex.MatchString(update.Message.Text) {
			log.Printf("bad word detected in chat %d from user %d", update.Message.Chat.ID, update.Message.From.ID)

			_, err := bot.SendMessage(ctx, tu.Message(
				tu.ID(update.Message.Chat.ID),
				"Не ругайся!",
			))
			if err != nil {
				log.Printf("failed to send warning message: %v", err)
				return nil
			}
		}

		return nil
	})

	// Stop handling
	defer bh.Stop()

	// Start handling
	log.Println("bot handler started")
	if err := bh.Start(); err != nil {
		log.Fatalf("bot handler error: %v", err)
	}
}

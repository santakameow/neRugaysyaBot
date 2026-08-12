package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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

	// create new bot
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		panic(err)
	}

	// get updates channel
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		panic(err)
	}

	// bot handler
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	// get bad words list from badWords.txt
	badWords, err := os.ReadFile("badWords.txt")
	if err != nil {
		log.Fatal("badWords.txt not found")
	}

	badWordsMap := make(map[string]struct{})

	for _, line := range strings.Split(string(badWords), "\n") {
		line = strings.TrimSpace(line)

		// skip empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		badWordsMap[strings.ToLower(line)] = struct{}{}
	}

	// main handler
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// get user message from updates
		userMessage := strings.ToLower(update.Message.Text)

		userWords := strings.Fields(strings.ToLower(userMessage))
		
		for _, word := range userWords {
			if _, exists := badWordsMap[word]; exists {
				// send message
				bot.SendMessage(ctx, tu.Message(
					tu.ID(update.Message.Chat.ID),
					"Не ругайся!",
				))
				return nil
			}
		}
		return nil
	})

	// Stop handling
	defer bh.Stop()

	// Start handling
	_ = bh.Start()
}

package main

import (
	"context"
	"fmt"
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

	var words []string

	for _, line := range strings.Split(string(badWords), "\n") {
		line = strings.TrimSpace(line)

		// skip empty lines
		if line == "" {
			continue
		}

		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		words = append(words, strings.ToLower(line))
	}

	// main handler
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// get user message from updates
		userMessage := strings.ToLower(update.Message.Text)

		for _, word := range words {
			// if string contains bad word
			if strings.Contains(userMessage, word) {
				// send message
				bot.SendMessage(ctx, tu.Message(
					// to last chat id
					tu.ID(update.Message.Chat.ID),
					// with this warn
					"Не ругайся!",
				))
				fmt.Println("bot sended message")
				return nil
			}
		}
		return nil
	})

	// Stop handling updates
	defer bh.Stop()

	// Start handling updates
	_ = bh.Start()
}

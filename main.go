package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	// context and correct exit
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	botToken := os.Getenv("TOKEN")

	// create new bot without debug logger (debug logger may expose sensitive info)
	bot, err := telego.NewBot(botToken)
	if err != nil {
		panic(err)
	}
	fmt.Println("bot started")

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
		panic(err)
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
				return nil
			}
		}
		return nil
	})

	// Stop handling updates
	defer func() { _ = bh.Stop() }()

	// Start handling updates
	_ = bh.Start()
}

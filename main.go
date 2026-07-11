package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	ctx := context.Background()
	botToken := os.Getenv("TOKEN")

	// Note: Please keep in mind that default logger may expose sensitive information, use in development only
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get updates channel
	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	// Create bot handler and specify from where to get updates
	bh, _ := th.NewBotHandler(bot, updates)

	// bothandler
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot.SendMessage(ctx, tu.Message(tu.ID(update.Message.Chat.ID), "meow"))
		return nil
	})

	// Stop handling updates
	defer func() { _ = bh.Stop() }()

	// Start handling updates
	_ = bh.Start()
}

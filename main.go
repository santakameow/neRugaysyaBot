package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// start bot with specified token
func startBot(botToken string) error {
	ctx := context.Background()

	// register new bot
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Printf("error: %s", err)
		os.Exit(1)
	}

	// get updates
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		fmt.Printf("error: %s", err)
		os.Exit(1)
	}

	// register bot handler
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		fmt.Printf("error: %s", err)
		os.Exit(1)
	}

	// handle all messages
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// send message if bad word detected
		if IsProfane(update.Message.Text) {
			bot.SendMessage(
				ctx,
				tu.Messagef(
					tu.ID(update.Message.Chat.ID),
					"%s, не ругайся!", update.Message.From.FirstName,
				).WithReplyParameters(&telego.ReplyParameters{
					MessageID: update.Message.MessageID,
				}))
		}
		return nil
	})

	defer func() {
		_ = bh.Stop()
	}()

	bh.Start()

	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error: %s", err)
	}

	// token of bot that comes from env
	botToken := os.Getenv("BOT_TOKEN")

	startBot(botToken)
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// start telegram bot with specified token
func startBot(botToken string, db *sql.DB) error {
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

	// handle any messages
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// send message if bad word detected
		if IsProfane(update.Message.Text) {
			err := incrementSwearCount(db, update.Message.From.ID)
			if err != nil {
				fmt.Printf("failed to increment swear count: %s\n", err)
			}
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
	}, th.AnyMessage())

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

	// token of bot that comes from env.
	// by default not set, that causes issues
	botToken := os.Getenv("BOT_TOKEN")

	// path to database. by default points to /data/stats.db
	// for development run i recommend to change DB_PATH in .env
	dbPath := os.Getenv("DB_PATH")

	db, err := InitDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	startBot(botToken, db)
}

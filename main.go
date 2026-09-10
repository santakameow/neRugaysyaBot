package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

const replyChance = 0.67

// start telegram bot with specified token
func startBot(botToken string, db *sql.DB) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	// handle statistics
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		StatsRequestsTotal.Inc()
		if update.Message == nil {
			return nil
		}
		swears, err := getSwearCount(db, update.Message.From.ID)
		if err != nil {
			DBErrorsTotal.Inc()
			return fmt.Errorf("failed to get swear count: %s", err)
		}

		bot.SendMessage(
			ctx,
			tu.Messagef(
				tu.ID(update.Message.Chat.ID),
				"%s, your swear count is: %d", update.Message.From.FirstName, swears,
			),
		)
		return nil
	}, th.CommandEqual("stats"))

	// handle any messages
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if update.Message == nil {
			return nil
		}
		MessagesTotal.Inc()
		// send message if bad word detected
		if IsProfane(update.Message.Text) {
			ProfaneTotal.Inc()
			fmt.Printf("profane message from %d: %q hits=%v\n", update.Message.From.ID, update.Message.Text, FindProfanity(update.Message.Text))
			err := incrementSwearCount(db, update.Message.From.ID)
			if err != nil {
				DBErrorsTotal.Inc()
				fmt.Printf("failed to increment swear count: %s\n", err)
			}

			if rand.Float32() > replyChance {
				fmt.Println("reply skipped by chance")
				return nil
			}

			_, err = bot.SendMessage(
				ctx,
				tu.Messagef(
					tu.ID(update.Message.Chat.ID),
					"%s, не ругайся!", update.Message.From.FirstName,
				).WithReplyParameters(&telego.ReplyParameters{
					MessageID: update.Message.MessageID,
				}))
			if err != nil {
				fmt.Printf("failed to send reply: %s\n", err)
			} else {
				RepliesTotal.Inc()
			}
		}
		return nil
	}, th.AnyMessage())

	// Initialize done chan
	done := make(chan struct{}, 1)

	// Handle stop signal (Ctrl+C)
	go func() {
		// Wait for stop signal
		<-ctx.Done()
		fmt.Println("Stopping...")

		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*20)
		defer stopCancel()

	loop:
		for len(updates) > 0 {
			select {
			case <-stopCtx.Done():
				break loop
			case <-time.After(time.Microsecond * 100):
				// Continue
			}
		}
		fmt.Println("Long polling done")

		_ = bh.StopWithContext(stopCtx)
		fmt.Println("Bot handler done")

		// Notify that stop is done
		done <- struct{}{}
	}()

	// Start handling in goroutine
	go func() { _ = bh.Start() }()
	fmt.Println("Handling updates...")

	// Wait for the stop process to be completed
	<-done
	fmt.Println("Done")

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

	metricsAddr := os.Getenv("METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":2112"
	}
	go startMetricsServer(metricsAddr)
	fmt.Printf("metrics listening on %s\n", metricsAddr)

	db, err := InitDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	startBot(botToken, db)
}

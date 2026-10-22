package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

const replyChance = 0.67

func init() {
	// Telego's default JSON encoder can allocate gigabytes when concurrent
	// Update.Clone calls compile the same encoder during startup catch-up.
	// Configure this before any bot goroutines start.
	telego.SetJSONMarshal(json.Marshal)
	telego.SetJSONUnmarshal(json.Unmarshal)
}

// start telegram bot with specified token
func startBot(botToken string, db *sql.DB) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// register new bot
	bot, err := telego.NewBot(botToken)
	if err != nil {
		return err
	}

	// get updates
	updates, err := bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Limit:   20,
		Timeout: 30,
	}, telego.WithLongPollingBuffer(4))
	if err != nil {
		return err
	}

	// Use the routing group directly: BotHandler.Start starts an unbounded
	// goroutine per update, defeating backpressure during backlog replay.
	bh := &th.HandlerGroup{}

	// handle statistics
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		StatsRequestsTotal.Inc()
		if update.Message == nil {
			return nil
		}
		swears, err := getSwearCount(ctx, db, update.Message.From.ID)
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
			err := incrementSwearCount(ctx, db, update.Message.From.ID)
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

	fmt.Println("Handling updates...")
	runUpdateWorkers(ctx, updates, func(ctx context.Context, update telego.Update) {
		if err := bh.HandleUpdate(ctx, bot, update); err != nil {
			fmt.Printf("failed to handle update %d: %s\n", update.UpdateID, err)
		}
	})
	fmt.Println("Done")
	return nil
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

	// set metrics address from env.
	// if not set, use default :2112
	metricsAddr := os.Getenv("METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":2112"
	}
	// start metric server
	go startMetricsServer(metricsAddr)
	fmt.Printf("metrics listening on %s\n", metricsAddr)

	db, err := InitDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := startBot(botToken, db); err != nil {
		fmt.Printf("bot stopped: %s\n", err)
	}
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// start bot with specified token
func startBot(botToken string, patterns []string) error {
	ctx := context.Background()

	fmt.Println(patterns)

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

	words := patterns

	// handle all messages
	bh.Handle(func(ctx *th.Context, update telego.Update) error {

		// send message if bad word detected
		for _, word := range words {
			if strings.Contains(strings.ToLower(update.Message.Text), word) {
				bot.SendMessage(
					ctx,
					tu.Messagef(
						tu.ID(update.Message.Chat.ID),
						"%s, не ругайся!", update.Message.From.FirstName,
					).WithReplyParameters(&telego.ReplyParameters{
					MessageID: update.Message.MessageID,
				}))
			}
		}
		
		return nil
	})

	defer func ()  {
		_ = bh.Stop()
	} ()

	bh.Start()

	return err
}

// load patterns from specified file
func loadPatterns(file string) []string {

	badWords, _ := os.Open(file)
	defer badWords.Close()

	fileScanner := bufio.NewScanner(badWords)

	var patterns []string

	for fileScanner.Scan() {
		if fileScanner.Text() == "" || strings.HasPrefix(fileScanner.Text(), "#") {
			continue
		}
		patterns = append(patterns, fileScanner.Text())
	}
	if err := fileScanner.Err(); err != nil {
		fmt.Println(err)
	}

	return patterns
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error: %s", err)
	}

	botToken := os.Getenv("BOT_TOKEN")

	patterns := loadPatterns("badWords.txt")


	startBot(botToken, patterns)
}
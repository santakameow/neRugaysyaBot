package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mymmrac/telego"
)

func main() {
	// get bot token from env
	botToken := os.Getenv("BOT_TOKEN")

	// create bot
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// call method getMe
	botUser, err := bot.GetMe(context.Background())
	if err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Printf("bot user: %+v\n", botUser)
}

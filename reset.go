package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
)

func handleResetCount(ctx context.Context, db *sql.DB, ownerID int64, message *telego.Message) (string, error) {
	// Anonymous admins and channel senders must not inherit owner privileges.
	if ownerID <= 0 || message == nil || message.From == nil || message.SenderChat != nil || message.From.ID != ownerID {
		return "", nil
	}

	args := strings.Fields(message.Text)
	if len(args) != 2 {
		return "Usage: /resetcount <user_id>", nil
	}
	userID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || userID <= 0 {
		return "Please provide a valid positive Telegram user ID.", nil
	}
	if err := resetSwearCount(ctx, db, userID); err != nil {
		return "", err
	}
	return fmt.Sprintf("Swear count reset to 0 for user %d.", userID), nil
}

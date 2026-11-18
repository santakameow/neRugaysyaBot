package main

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"strings"

	"github.com/mymmrac/telego"
)

type swearEntry struct {
	userID int64
	count  int
	name   string
}

// Track only directly observed users, not forwarded messages or reply authors.
// This does not create swear-count entries for users without statistics.
func trackChatUsers(ctx context.Context, db *sql.DB, message *telego.Message) error {
	if message == nil || (message.Chat.Type != "group" && message.Chat.Type != "supergroup") {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	users := append([]telego.User(nil), message.NewChatMembers...)
	if message.From != nil && message.SenderChat == nil {
		users = append(users, *message.From)
	}
	for _, user := range users {
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO chat_users (chat_id, user_id, name, is_bot) VALUES (?, ?, ?, ?)
			ON CONFLICT(chat_id, user_id) DO UPDATE SET name = excluded.name, is_bot = excluded.is_bot
		`, message.Chat.ID, user.ID, name, user.IsBot)
		if err != nil {
			return err
		}
	}
	if message.LeftChatMember != nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM chat_users WHERE chat_id = ? AND user_id = ?`, message.Chat.ID, message.LeftChatMember.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func getSwearLeaderboard(ctx context.Context, db *sql.DB, chatID int64) ([]swearEntry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT u.user_id, u.swear_count, c.name
		FROM users u JOIN chat_users c ON c.user_id = u.user_id
		WHERE c.chat_id = ? AND c.is_bot = 0
		ORDER BY u.swear_count ASC, u.user_id ASC`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []swearEntry
	for rows.Next() {
		var entry swearEntry
		if err := rows.Scan(&entry.userID, &entry.count, &entry.name); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func globalStatsMessages(entries []swearEntry) []string {
	const heading = "🏆 <b>Top in chat:</b>\n\n"
	// Bound even the encoded HTML to comfortably fit Telegram's 4096-character limit.
	const maxLength = 4000
	var messages []string
	message := heading
	for i, entry := range entries {
		name := entry.name
		if name == "" {
			name = fmt.Sprintf("User %d", entry.userID)
		}
		line := fmt.Sprintf("%d. <a href=\"tg://user?id=%d\">%s</a> — %d swears\n", i+1, entry.userID, html.EscapeString(name), entry.count)
		if len(message)+len(line) > maxLength {
			messages = append(messages, message)
			message = heading
		}
		message += line
	}
	if len(entries) == 0 {
		return []string{"No observed chat members with swear statistics yet."}
	}
	return append(messages, message)
}

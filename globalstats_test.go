package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mymmrac/telego"
)

func TestObservedGlobalStats(t *testing.T) {
	ctx := context.Background()
	db, err := InitDB(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Existing stats survive the schema migration/restart.
	for id, count := range map[int64]int{1: 3, 2: 0, 3: 10, 4: 9, 5: 8, 6: 0} {
		if _, err := db.Exec(`INSERT INTO users (user_id, swear_count) VALUES (?, ?)`, id, count); err != nil {
			t.Fatal(err)
		}
	}
	track := func(message *telego.Message) {
		t.Helper()
		if err := trackChatUsers(ctx, db, message); err != nil {
			t.Fatal(err)
		}
	}
	chat := telego.Chat{ID: -123, Type: "supergroup"}
	for _, user := range []telego.User{
		{ID: 1, FirstName: "Old name"},
		{ID: 2, FirstName: "Bob"},
		{ID: 3, FirstName: "Bot", IsBot: true},
		{ID: 4, FirstName: "Leaving"},
		{ID: 7, FirstName: "No stats"},
	} {
		track(&telego.Message{Chat: chat, From: &user})
	}
	track(&telego.Message{Chat: chat, From: &telego.User{ID: 1, FirstName: "<Alice>", LastName: "& Co"}})
	track(&telego.Message{Chat: chat, LeftChatMember: &telego.User{ID: 4}})
	track(&telego.Message{Chat: telego.Chat{ID: -456, Type: "group"}, From: &telego.User{ID: 5}})
	// Anonymous admin messages must not introduce a fake user identity.
	track(&telego.Message{Chat: chat, From: &telego.User{ID: 6}, SenderChat: &chat})
	track(nil)
	entries, err := getSwearLeaderboard(ctx, db, chat.ID)
	if err != nil {
		t.Fatal(err)
	}
	messages := globalStatsMessages(entries)
	want := "🏆 <b>Top in chat:</b>\n\n" +
		"1. <a href=\"tg://user?id=2\">Bob</a> — 0 swears\n" +
		"2. <a href=\"tg://user?id=1\">&lt;Alice&gt; &amp; Co</a> — 3 swears\n"
	if len(messages) != 1 || messages[0] != want {
		t.Fatalf("unexpected messages: %q", messages)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE user_id = 7`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("tracking created statistics: count=%d err=%v", count, err)
	}
	// A join notification restores a user whose departure was observed.
	track(&telego.Message{Chat: chat, NewChatMembers: []telego.User{{ID: 4, FirstName: "Rejoined"}}})
	entries, err = getSwearLeaderboard(ctx, db, chat.ID)
	if err != nil || len(entries) != 3 || entries[2].userID != 4 {
		t.Fatalf("rejoin not tracked: %v err=%v", entries, err)
	}
}

func TestGlobalStatsFormatting(t *testing.T) {
	if messages := globalStatsMessages(nil); len(messages) != 1 || !strings.Contains(messages[0], "No observed") {
		t.Fatalf("unexpected empty response: %q", messages)
	}
	var entries []swearEntry
	for id := int64(1); id <= 200; id++ {
		entries = append(entries, swearEntry{userID: id, name: strings.Repeat("A&", 32)})
	}
	messages := globalStatsMessages(entries)
	if len(messages) < 2 {
		t.Fatal("expected multiple messages")
	}
	for _, message := range messages {
		if len(message) > 4000 || !strings.HasSuffix(message, "swears\n") {
			t.Fatalf("invalid message: %q", message)
		}
	}
	if !strings.Contains(messages[len(messages)-1], "200. <a") {
		t.Fatal("ranking was truncated or numbering restarted")
	}
	if message := globalStatsMessages([]swearEntry{{userID: 42}})[0]; !strings.Contains(message, ">User 42</a>") {
		t.Fatal("missing fallback name")
	}
}

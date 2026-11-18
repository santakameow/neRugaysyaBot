package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mymmrac/telego"
)

func TestResetCount(t *testing.T) {
	ctx := context.Background()
	db, err := InitDB(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const ownerID int64 = 123456789
	for _, tc := range []struct {
		name      string
		owner     int64
		message   *telego.Message
		wantReply string
		wantCount int
	}{
		{"unauthorized", ownerID, &telego.Message{From: &telego.User{ID: 2}, Text: "/resetcount 42"}, "", 1},
		{"disabled", 0, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount 42"}, "", 1},
		{"negative owner", -1, &telego.Message{From: &telego.User{ID: -1}, Text: "/resetcount 42"}, "", 1},
		{"nil message", ownerID, nil, "", 1},
		{"nil sender", ownerID, &telego.Message{Text: "/resetcount 42"}, "", 1},
		{"sender chat", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, SenderChat: &telego.Chat{ID: -123}, Text: "/resetcount 42"}, "", 1},
		{"missing argument", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount"}, "Usage:", 1},
		{"extra argument", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount 42 43"}, "Usage:", 1},
		{"invalid id", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount abc"}, "Please provide", 1},
		{"zero id", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount 0"}, "Please provide", 1},
		{"negative id", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount -42"}, "Please provide", 1},
		{"overflow", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount 99999999999999999999"}, "Please provide", 1},
		{"reset", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount 42"}, "Swear count reset", 0},
		{"bot mention", ownerID, &telego.Message{From: &telego.User{ID: ownerID}, Text: "/resetcount@mybot 42"}, "Swear count reset", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := resetSwearCount(ctx, db, 42); err != nil {
				t.Fatal(err)
			}
			for _, id := range []int64{42, 43} {
				if err := incrementSwearCount(ctx, db, id); err != nil {
					t.Fatal(err)
				}
			}
			before, err := getSwearCount(ctx, db, 43)
			if err != nil {
				t.Fatal(err)
			}
			reply, err := handleResetCount(ctx, db, tc.owner, tc.message)
			if err != nil {
				t.Fatal(err)
			}
			if (tc.wantReply == "" && reply != "") || !strings.HasPrefix(reply, tc.wantReply) {
				t.Fatalf("unexpected reply: %q", reply)
			}
			count, err := getSwearCount(ctx, db, 42)
			if err != nil || count != tc.wantCount {
				t.Fatalf("count=%d err=%v", count, err)
			}
			after, err := getSwearCount(ctx, db, 43)
			if err != nil || after != before {
				t.Fatalf("other user changed: %d -> %d, err=%v", before, after, err)
			}
		})
	}

	// Unknown users can be reset repeatedly and incremented normally afterward.
	for range 2 {
		if err := resetSwearCount(ctx, db, 999); err != nil {
			t.Fatal(err)
		}
	}
	if err := incrementSwearCount(ctx, db, 999); err != nil {
		t.Fatal(err)
	}
	if count, err := getSwearCount(ctx, db, 999); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

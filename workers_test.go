package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mymmrac/telego"
)

func waitForWorkers(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("workers did not finish")
	}
}

func TestUpdateWorkersBoundConcurrencyAndProcessBacklog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan telego.Update, 488)
	for i := range cap(updates) {
		updates <- telego.Update{UpdateID: i}
	}
	close(updates)
	started := make(chan struct{}, cap(updates))
	release := make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(release) })
	var active, peak, processed atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		runUpdateWorkers(ctx, updates, func(ctx context.Context, update telego.Update) {
			n := active.Add(1)
			for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
			}
			started <- struct{}{}
			select {
			case <-release:
			case <-ctx.Done():
			}
			active.Add(-1)
			processed.Add(1)
		})
	}()
	for range updateWorkers {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("worker did not start")
		}
	}
	select {
	case <-started:
		t.Fatal("concurrency exceeded worker limit")
	case <-time.After(50 * time.Millisecond):
	}
	unblock.Do(func() { close(release) })
	waitForWorkers(t, done)
	if peak.Load() != updateWorkers || processed.Load() != 488 {
		t.Fatalf("peak=%d processed=%d", peak.Load(), processed.Load())
	}
}

func TestUpdateWorkersCancelInFlightAndPendingUpdates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan telego.Update, 100)
	for range cap(updates) {
		updates <- telego.Update{}
	}
	started := make(chan context.Context, updateWorkers)
	var processed atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		runUpdateWorkers(ctx, updates, func(ctx context.Context, update telego.Update) {
			processed.Add(1)
			started <- ctx
			<-ctx.Done()
		})
	}()
	for range updateWorkers {
		select {
		case handlerCtx := <-started:
			deadline, ok := handlerCtx.Deadline()
			if !ok || time.Until(deadline) > updateTimeout {
				t.Fatal("handler must have a bounded deadline")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("worker did not start")
		}
	}
	cancel()
	waitForWorkers(t, done)
	if processed.Load() != updateWorkers {
		t.Fatalf("started pending work after cancellation: %d", processed.Load())
	}
}

func TestUpdateWorkersStopWhenIdle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan telego.Update)
	done := make(chan struct{})
	go func() {
		defer close(done)
		runUpdateWorkers(ctx, updates, func(context.Context, telego.Update) {})
	}()
	cancel()
	waitForWorkers(t, done)
}

func TestConcurrentUpdateClones(t *testing.T) {
	// Telego clones updates before matching predicates. Exercise cold-start
	// encoder initialization concurrently, like a backlog received on startup.
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			original := telego.Update{UpdateID: i, Message: &telego.Message{
				MessageID: i, Text: "test", Chat: telego.Chat{ID: 1, Type: "group"},
			}}
			cloned, err := original.CloneSafe()
			if err != nil {
				t.Error(err)
				return
			}
			if cloned.Message == original.Message || cloned.Message.Text != original.Message.Text {
				t.Error("update must be deep-copied without changing its text")
			}
		})
	}
	wg.Wait()
}

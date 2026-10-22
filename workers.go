package main

import (
	"context"
	"sync"
	"time"

	"github.com/mymmrac/telego"
)

const (
	updateWorkers = 4
	updateTimeout = 20 * time.Second
)

// A fixed worker pool propagates backpressure to long polling instead of
// retaining an unbounded number of updates and handler goroutines.
func runUpdateWorkers(ctx context.Context, updates <-chan telego.Update, handle func(context.Context, telego.Update)) {
	var workers sync.WaitGroup
	for range updateWorkers {
		workers.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case update, ok := <-updates:
					if !ok || ctx.Err() != nil {
						return
					}
					updateCtx, cancel := context.WithTimeout(ctx, updateTimeout)
					handle(updateCtx, update.WithContext(updateCtx))
					cancel()
				}
			}
		})
	}
	workers.Wait()
}

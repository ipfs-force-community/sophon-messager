package service

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestJoinGroup_WaitsForLoopsToDrain pins that shutdown returns as soon as the
// loops have stopped: a join that returns before the group drained, or one that
// only returns when the safety-net timeout expires, must fail. Failing values:
// elapsed == 0 for a no-op join, elapsed >= the timeout for a timeout-bound wait.
func TestJoinGroup_WaitsForLoopsToDrain(t *testing.T) {
	old := shutdownTimeout
	shutdownTimeout = 5 * time.Second
	defer func() { shutdownTimeout = old }()

	// Keeps the group pending while the join runs, so the ordering assertion below
	// has a window in which a join that returns early cannot see it drained.
	const drainDelay = 20 * time.Millisecond

	var wg sync.WaitGroup
	wg.Add(1)
	drained := make(chan struct{})
	go func() {
		defer wg.Done()
		time.Sleep(drainDelay)
		close(drained)
	}()

	start := time.Now()
	joinGroup(context.Background(), "test", &wg)
	elapsed := time.Since(start)

	select {
	case <-drained:
	default:
		t.Fatal("joinGroup returned before the group was drained")
	}
	if elapsed > time.Second {
		t.Fatalf("joinGroup waited out the %s timeout (drain took %s) instead of returning once the group drained",
			shutdownTimeout, elapsed)
	}
}

// TestJoinGroup_GivesUpWhenLoopsIgnoreCancellation pins the safety net: a loop
// that ignores cancellation must not block shutdown forever, and giving up is
// not an error. Failing values: elapsed == 0 for a join that never waits,
// elapsed >> the timeout for one that waits unconditionally.
func TestJoinGroup_GivesUpWhenLoopsIgnoreCancellation(t *testing.T) {
	old := shutdownTimeout
	shutdownTimeout = 50 * time.Millisecond
	defer func() { shutdownTimeout = old }()

	var wg sync.WaitGroup
	wg.Add(1) // never released

	start := time.Now()
	joinGroup(context.Background(), "test", &wg)
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Fatalf("gave up before the safety-net timeout elapsed: %s", elapsed)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("did not give up after the timeout: %s", elapsed)
	}
}

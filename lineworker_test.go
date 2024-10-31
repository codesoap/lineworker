package lineworker_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/codesoap/lineworker"
)

// TestStop tests if an external stopping of the worker pool works
// without problems.
func TestStop(t *testing.T) {
	pool := lineworker.NewWorkerPool(2, func(a any) (any, error) { return 0, nil })
	stoppedProcessing := &atomic.Bool{}
	go func() {
		for {
			if ok := pool.Process(0); !ok {
				break
			}
		}
		stoppedProcessing.Store(true)
	}()
	for i := 0; i < 10; i++ {
		_, err := pool.Next()
		if err != nil {
			t.Error("Next returned error:", err)
			break
		}
	}
	pool.Stop()
	pool.DiscardWork()
	for i := 0; i < 10; i++ {
		if stoppedProcessing.Load() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("Processing did not stop after call to Stop.")
}

// TestStopTwice tests if stopping a worker pool twice works.
func TestStopTwice(t *testing.T) {
	pool := lineworker.NewWorkerPool(1, func(a any) (any, error) { return 42, nil })
	if ok := pool.Process(0); !ok {
		t.Error("could not process work")
	}
	pool.Stop()
	pool.Stop()
	res, err := pool.Next()
	if err != nil {
		t.Errorf("unexpected error when retrieving result: %v", err)
	}
	if res != 42 {
		t.Errorf("got wrong result (expected 42): %d", res)
	}
	_, err = pool.Next()
	if err != lineworker.EOS {
		t.Error("expected EOS, but did not occur")
	}
}

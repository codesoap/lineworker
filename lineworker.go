// package lineworker provides a worker pool with a fixed amount of
// workers. It outputs work results in the order the work was given. The
// package is designed for serial data input and output; the functions
// Process and Next must never be called in parallel.
//
// Each worker caches at most one result, so that no new work is
// processed, if as many results are waiting to be consumed as there are
// workers.
package lineworker

import (
	"fmt"
	"sync/atomic"
)

// EOS is the error returned by Next when no more results are available.
var EOS = fmt.Errorf("no more results available")

type WorkFunc[IN, OUT any] func(in IN) (OUT, error)

type WorkerPool[IN, OUT any] struct {
	workFunc     WorkFunc[IN, OUT]
	processCalls int
	nextCalls    int
	work         []chan IN
	out          []chan workResult[OUT]
	stopped      atomic.Bool
}

// NewWorkerPool creates a new worker pool with workerCount workers
// waiting to process data of type IN to results of type OUT via f.
func NewWorkerPool[IN, OUT any](workerCount int, f WorkFunc[IN, OUT]) *WorkerPool[IN, OUT] {
	pool := WorkerPool[IN, OUT]{
		workFunc: f,
		work:     make([]chan IN, workerCount),
		out:      make([]chan workResult[OUT], workerCount),
	}
	for i := 0; i < workerCount; i++ {
		pool.work[i] = make(chan IN)
		pool.out[i] = make(chan workResult[OUT])
		go func() {
			for {
				if w, ok := <-pool.work[i]; ok {
					out, err := pool.workFunc(w)
					pool.out[i] <- workResult[OUT]{result: out, err: err}
				} else {
					close(pool.out[i])
					return
				}
			}
		}()
	}
	return &pool
}

// Process queues a new input for processing. If all workers are
// currently busy, Process will block.
//
// Process will return true if the input has been accepted. If Stop has
// been called previously, Process will discard the given input and
// return false.
func (w *WorkerPool[IN, OUT]) Process(input IN) bool {
	if w.stopped.Load() {
		return false
	}
	w.work[w.processCalls%len(w.work)] <- input
	w.processCalls++
	return true
}

// Next will return the next result with its error. If the next result
// is not yet ready, it will block. If no more results are available,
// the EOS error will be returned.
func (w *WorkerPool[IN, OUT]) Next() (OUT, error) {
	res, ok := <-w.out[w.nextCalls%len(w.out)]
	if !ok {
		return *new(OUT), EOS
	}
	w.nextCalls++
	return res.result, res.err
}

// Stop should be called after all calls to Process have been made. It
// stops the workers from accepting new work and allows their resources
// to be released after all results have been consumed via Next or
// discarded with DiscardWork.
//
// Further calls to Stop after the first call will do nothing.
func (w *WorkerPool[IN, OUT]) Stop() {
	if !w.stopped.Swap(true) {
		for _, work := range w.work {
			close(work)
		}
	}
}

// DiscardWork recieves and discards all pending work results, so that
// workers can quit after Stop has been called. It will block until all
// workers have quit.
//
// DiscardWork must only be called after Stop has been called.
func (w *WorkerPool[IN, OUT]) DiscardWork() {
	for _, out := range w.out {
		for {
			if _, ok := <-out; !ok {
				break
			}
		}
	}
}

type workResult[OUT any] struct {
	result OUT
	err    error
}

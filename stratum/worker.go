package stratum

import (
	"sync"
)

type Worker struct {
	lock     sync.Mutex
	// endpoint *birpc.Endpoint
	timeoutCh   chan bool
	pool  *Pool
}

func NewWorker(pool *Pool) *Worker {
	worker := &Worker{
		timeoutCh: make(chan bool, 1),
		pool: pool,
	}
	return worker
}

func (w *Worker) Close() {
	// nothing here
}

package stratum

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Pool struct {
	lock     sync.Mutex
	address  string
	order    *Order
	upstream *StratumClient
	workers  map[*Worker]bool
	jobs     map[string]*Job
	active   bool
	stable   bool
	closing  bool

	nonceCounter NonceCounter
}

func NewPool(order *Order, errch chan error) (pool *Pool, err error) {
	conn, err := net.Dial("tcp", order.Address())
	if err != nil {
		return nil, err
	}

	upstream := NewClient(conn, errch)
	err = upstream.Subscribe()
	if err != nil {
		return nil, err
	}

	err = upstream.Authorize(order.Username, order.Password)
	if err != nil {
		return nil, err
	}

	order.markConnected()
	return NewPoolWithConn(order, upstream)
}

func NewPoolWithConn(order *Order, upstream *StratumClient) (*Pool, error) {
	context := upstream.Context()
	if context.ExtraNonce2Size != ExtraNonce2Size+ExtraNonce3Size {
		errmsg := fmt.Sprintf("Invalid nonce sizes, must add up to %d", context.ExtraNonce2Size)
		return nil, errors.New(errmsg)
	}

	p := &Pool{
		address:  order.Address(),
		order:    order,
		upstream: upstream,
		workers:  make(map[*Worker]bool),
	}

	p.nonceCounter = NewProxyExtraNonceCounter(context.ExtraNonce1, ExtraNonce2Size, ExtraNonce3Size)

	go p.Serve(DefaultPoolTimeout)

	return p, nil
}

func (p *Pool) Serve(timeout time.Duration) {
	for {
		if p.isClosed() {
			break
		}
		ctx := p.Context()
		select {
		case job := <-ctx.JobCh:
			p.onNewJob(&job)
		case _ = <-ctx.ShutdownCh:
			p.Shutdown()
			break
		case <-time.After(timeout):
			log.Printf("Pool %s timeout in %.1f minutes.", p.address, timeout.Minutes())
			p.Shutdown()
			break
		}
	}

	log.Printf("Pool %s stop serving.", p.address)
}

func (p *Pool) Context() *ClientContext {
	if p.upstream == nil {
		return nil
	}
	return p.upstream.Context()
}

func (p *Pool) isClosed() bool {
	if p.upstream == nil {
		return true
	}

	return p.closing
}

func (p *Pool) Shutdown() {
	if p.upstream == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.active = false
	p.stable = false
	p.closing = true

	log.Printf("Stopping pool %s...", p.address)

	p.upstream.Close()
	p.upstream = nil

	log.Printf("Pool %s stopped.", p.address)

	// relocate miners
}

func (p *Pool) addWorker(worker *Worker) {
	p.workers[worker] = true
}

func (p *Pool) removeWorker(worker *Worker) {
	_, ok := p.workers[worker]
	if !ok {
		log.Printf("Work not found in pool %s.", p.address)
		return
	}
	delete(p.workers, worker)
}

func (p *Pool) nextNonce1() string {
	return p.nonceCounter.Next()
}

func (p *Pool) nonce2Size() int {
	return p.nonceCounter.Nonce2Size()
}

func (p *Pool) CurrentJob() (*Job, error) {
	ctx := p.Context()
	if ctx == nil {
		return nil, errors.New("Closed upstream.")
	}
	return p.Context().CurrentJob, nil
}

func (p *Pool) onNewJob(job *Job) {
	if job.CleanJobs {

	}
}

// broadcast mining jobs
func (p *Pool) broadcast(job *Job) {
	for worker, _ := range p.workers {
		worker.sendJob(job)
	}
	log.Printf("Broadcast job from %s to %d workers.", p.address, len(p.workers))
}

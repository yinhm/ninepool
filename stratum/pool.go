package stratum

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"sync"
	"time"
)

type Pool struct {
	lock       sync.RWMutex
	id         uint64
	address    string
	order      *Order
	upstream   *StratumClient
	workers    map[*Worker]bool
	jobs       map[string]*Job
	CurrentJob *Job
	active     bool
	stable     bool
	closing    bool

	nonceCounter NonceCounter
}

func NewPool(server *StratumServer, order *Order, errch chan error) (pool *Pool, err error) {
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

	return NewPoolWithConn(server, order, upstream, errch)
}

func FindPool(pid int) (*Pool, bool) {
	p, ok := DefaultServer.findPool(uint64(pid))
	return p, ok
}

func NewPoolWithConn(server *StratumServer, order *Order, upstream *StratumClient, errch chan error) (*Pool, error) {
	p := &Pool{
		id:      order.Id,
		address: order.Address(),
		order:   order,
		workers: make(map[*Worker]bool),
		jobs:    make(map[string]*Job),
	}
	err := p.setUpstream(upstream)
	if err != nil {
		return nil, err
	}

	go p.Serve(server, DefaultPoolTimeout, errch)

	order.markConnected()
	return p, nil
}

func (p *Pool) setUpstream(upstream *StratumClient) error {
	p.upstream = upstream
	context := upstream.Context()
	if context.ExtraNonce2Size != ExtraNonce2Size+ExtraNonce3Size {
		errmsg := fmt.Sprintf("Invalid nonce sizes, must add up to %d", context.ExtraNonce2Size)
		return errors.New(errmsg)
	}
	context.pid = p.id
	p.nonceCounter = NewProxyExtraNonceCounter(context.ExtraNonce1, ExtraNonce2Size, ExtraNonce3Size)
	p.active = true
	return nil
}

func (p *Pool) Serve(server *StratumServer, timeout time.Duration, errch chan error) {
	for {
		if p.isClosed() {
			break
		}
		ctx := p.Context()
		select {
		case job := <-ctx.JobCh:
			p.newJob(job)
		case err := <-errch:
			glog.Infof("Pool %s lost connection: %s, try reconnect...", p.address, err)
			err = p.reconnect(errch)
			if err != nil {
				glog.Infof("reconnect to %s failed, shutdown...", p.address)
				server.removePool(p)
				p.Shutdown()
			}
		case <-time.After(timeout):
			glog.Infof("Pool %s timeout in %.1f minutes.", p.address, timeout.Minutes())
			server.removePool(p)
			p.Shutdown()
			break
		}
	}

	glog.Infof("Pool %s stop serving.", p.address)
}

func (p *Pool) Context() *ClientContext {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.upstream == nil {
		return nil
	}
	return p.upstream.Context()
}

func (p *Pool) isClosed() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.upstream == nil {
		return true
	}
	return p.closing
}

func (p *Pool) isAvailable() bool {
	if p.isClosed() {
		return false
	}
	if !p.active {
		return false
	}
	if p.reachLimit() {
		return false
	}
	return true
}

// FIXME: reach limit when:
//  - no more nonce available
//  - reach limited ghs
func (p *Pool) reachLimit() bool {
	return false
}

func (p *Pool) reconnect(errch chan error) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	order := p.order
	order.markDead()
	p.upstream = nil
	p.active = false

	// close worker after upstream resetted, no more client can connect or
	// reconnect between the pool reconnection.
	p.closeWorkers()

	conn, err := net.Dial("tcp", order.Address())
	if err != nil {
		return err
	}

	upstream := NewClient(conn, errch)
	err = upstream.Subscribe()
	if err != nil {
		return err
	}

	err = upstream.Authorize(order.Username, order.Password)
	if err != nil {
		return err
	}

	order.markConnected()

	err = p.setUpstream(upstream)
	if err != nil {
		return err
	}
	glog.Infof("Pool %s reconnected.", p.address)
	return nil
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

	glog.Infof("Stopping pool %s...", p.address)

	p.upstream.Close()
	p.upstream = nil

	glog.Infof("Pool %s stopped.", p.address)

	// TODO: relocate miners
}

func (p *Pool) addWorker(worker *Worker) {
	p.lock.Lock()
	p.workers[worker] = true
	p.lock.Unlock()
}

func (p *Pool) removeWorker(worker *Worker) {
	_, ok := p.workers[worker]
	if !ok {
		glog.Infof("Work not found in pool %s.", p.address)
		return
	}
	delete(p.workers, worker)
}

func (p *Pool) closeWorkers() {
	// disconnect all workers
	glog.Infof("Closing %d workers.", len(p.workers))
	for worker, _ := range p.workers {
		worker.Close()
	}
}

func (p *Pool) nextNonce1() string {
	return p.nonceCounter.Next()
}

func (p *Pool) nonce2Size() int {
	return p.nonceCounter.Nonce2Size()
}

func (p *Pool) findJob(jobId string) (*Job, bool) {
	p.lock.RLock()
	job, ok := p.jobs[jobId]
	p.lock.RUnlock()

	return job, ok
}

func (p *Pool) newJob(job *Job) {
	p.lock.Lock()
	if job.CleanJobs {
		p.jobs = make(map[string]*Job)
	}
	p.jobs[job.JobId] = job
	p.CurrentJob = job
	p.lock.Unlock()
	go p.broadcast(job)
}

// broadcast mining jobs
func (p *Pool) broadcast(job *Job) {
	for worker, _ := range p.workers {
		worker.sendJob(job)
	}
	p.lock.RLock()
	count := len(p.workers)
	p.lock.RUnlock()
	glog.Infof("Broadcast job from %s to %d workers.", p.address, count)
}

// submit job to upstream
func (p *Pool) submit(diff float64, jobId, extraNonce1, extraNonce2, ntime, nonce, hash string) {
	ctx := p.Context()
	if ctx == nil {
		glog.Infof("share can not submit, lost connection to pool\n")
	}
	if diff/ctx.Difficulty < 0.99 {
		glog.Infof("[Pool] diff lower than the pool, no submit.")
		return
	}
	nonce2 := p.nonceCounter.Nonce1Suffix(extraNonce1) + extraNonce2
	err := p.upstream.Submit(ctx.Username, jobId, nonce2, ntime, nonce)
	// TODO: log upstream acceptence
	if err != nil {
		glog.Infof("[Pool] share rejected %s, %s.", hash, err.Error())
		return
	}
	glog.Infof("[Pool] share accepted: %s\n", hash)
}

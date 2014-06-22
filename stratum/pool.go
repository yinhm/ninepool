package stratum

import (
	"log"
	"net"
	"sync"
)

type Pool struct {
	lock     sync.Mutex
	address  string
	order    *Order
	upstream *StratumClient
	workers  map[*Worker]bool
	active   bool
	stable   bool
	closing  bool
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

	p := NewPoolWithConn(order, upstream)
	return p, nil
}

func NewPoolWithConn(order *Order, upstream *StratumClient) (pool *Pool) {
	return &Pool{
		address:  order.Address(),
		order:    order,
		upstream: upstream,
		workers:  make(map[*Worker]bool),
	}
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
		log.Printf("work not found in pool %s.", p.address)
    return
  }
	delete(p.workers, worker)
}

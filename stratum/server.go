package stratum

import (
	"errors"
	"github.com/golang/glog"
	"github.com/tv42/topic"
	"github.com/yinhm/birpc"
	"github.com/yinhm/birpc/jsonmsg"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
)

var ErrServerUnexpected = errors.New("Server error.")
var DefaultServer *StratumServer

type StratumServer struct {
	lock sync.Mutex
	*Stratum
	options Options
	workers map[*birpc.Endpoint]*Worker
	pools   map[uint64]*Pool
	orders  map[uint64]*Order
	errCh   chan error
	sigCh   chan os.Signal
	closing bool
}

func NewStratumServer(options Options) *StratumServer {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	DefaultServer = &StratumServer{
		Stratum: s,
		options: options,
		workers: make(map[*birpc.Endpoint]*Worker),
		pools:   make(map[uint64]*Pool),
		orders:  InitOrders("x11"),
		errCh:   make(chan error),
		sigCh:   make(chan os.Signal),
	}
	mining := &Mining{}
	// ss.registry.RegisterService(ss)
	DefaultServer.registry.RegisterService(mining)
	return DefaultServer
}

func (s *StratumServer) Start(l net.Listener) error {
	defer s.close()

	go s.startPools()
	go s.serve(l)

	signal.Notify(s.sigCh, os.Interrupt, os.Kill)

	// Block until a signal is received or we got an error
	select {
	case signal := <-s.sigCh:
		glog.Infof("Got signal %s, waiting for shutdown...", signal)
		s.Shutdown()
		return nil
	case err := <-s.errCh:
		glog.Infof("Server shutdown with error: %s", err)
		s.Shutdown()
		return err
	}
	return nil
}

func (s *StratumServer) serve(l net.Listener) {
	for {
		if s.closing == true {
			return
		}

		conn, err := l.Accept()
		if err != nil {
			glog.Infof("Error on accept connect.")
			continue
		}

		go s.ServeConn(conn)
	}
}

func (s *StratumServer) ServeConn(conn net.Conn) {
	defer conn.Close()

	endpoint := s.newEndpoint(conn)

	glog.Infof("Client connected: %v\n", conn.RemoteAddr())
	err := endpoint.Serve()
	if err != nil {
		if err == io.EOF {
			glog.Infof("Client disconnect: %v", conn.RemoteAddr())
		} else {
			glog.Infof("Error on %v", err)
		}
	}

	s.lock.Lock()
	glog.Infof("Deleting worker %v", conn.RemoteAddr())
	worker, _ := s.workers[endpoint]
	worker.Close()
	delete(s.workers, endpoint)
	s.lock.Unlock()
}

func (s *StratumServer) newEndpoint(conn net.Conn) *birpc.Endpoint {
	ep := birpc.NewEndpoint(jsonmsg.NewCodec(conn), s.registry)
	worker := NewWorker(ep, s.options.SubscribeTimeout)
	s.lock.Lock()
	s.workers[ep] = worker
	s.lock.Unlock()
	return ep
}

func (s *StratumServer) AddOrder(order *Order) {
	s.lock.Lock()
	s.orders[order.Id] = order
	s.lock.Unlock()
}

func (s *StratumServer) startPools() {
	for _, order := range s.orders {
		s.activeOrder(order)
	}
}

func (s *StratumServer) activeOrder(order *Order) {
	// test if actived
	_, ok := s.pools[order.Id]
	if ok {
		return
	}

	// connect to upstream pool
	glog.Infof("connecting to #%d, %s ...\n", order.Id, order.Address())

	errch := make(chan error, 1)
	pool, err := NewPool(s, order, errch)
	if err != nil {
		glog.Infof("[Pool #%d]: %s\n", order.Id, err.Error())
		order.markDead()
		return
	}

	s.ActivePool(order, pool, errch)
}

func (s *StratumServer) ActivePool(order *Order, pool *Pool, errch chan error) {
	s.lock.Lock()
	s.pools[order.Id] = pool
	s.lock.Unlock()
}

func (s *StratumServer) findPool(oid uint64) (*Pool, bool) {
	p, ok := s.pools[oid]
	return p, ok
}

func (s *StratumServer) Shutdown() {
	s.stopListen()
	// TODO: move stop worker to pool?
	s.stopWorkers()
	s.stopPools()
}

func (s *StratumServer) stopListen() {
	s.closing = true
}

func (s *StratumServer) stopPools() {
	for _, pool := range s.pools {
		s.removePool(pool)
		pool.Shutdown()
	}
}

func (s *StratumServer) removePool(p *Pool) {
	s.lock.Lock()
	if _, ok := s.pools[p.id]; ok {
		delete(s.pools, p.id)
	}
	s.lock.Unlock()
}

func (s *StratumServer) stopWorkers() {
	glog.Infof("Stopping %d workers.", len(s.workers))
	for _, worker := range s.workers {
		worker.Close() // close codec
	}
}

func (s *StratumServer) firstPool() (*Pool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, pool := range s.pools {
		if pool.isAvailable() {
			return pool, nil
		}
	}

	return nil, errors.New("No pool available.")
}

// for testing
func (s *StratumServer) FirstWorker() (*Worker, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, worker := range s.workers {
		return worker, nil
	}

	return nil, errors.New("No worker")
}

// func (s *StratumServer) Connection(e *birpc.Endpoint) (conn *Connection, err error) {
// 	conn, ok := s.connections[e]
// 	if !ok {
// 		e.Close()
// 		return nil, ErrServerUnexpected
// 	}

// 	return conn, nil
// }

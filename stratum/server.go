package stratum

import (
	"errors"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/birpc/jsonmsg"
	"io"
	"log"
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
	perrchs map[uint64]chan error // pool error chans
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
		perrchs: make(map[uint64]chan error),
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
		log.Printf("Got signal %s, waiting for shutdown...", signal)
		s.Shutdown()
		return nil
	case err := <-s.errCh:
		log.Printf("Server shutdown with error: %s", err)
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
			log.Printf("Error on accept connect.")
			continue
		}

		go s.ServeConn(conn)
	}
}

func (s *StratumServer) ServeConn(conn net.Conn) {
	defer conn.Close()

	endpoint := s.newEndpoint(conn)

	log.Printf("Client connected: %v\n", conn.RemoteAddr())
	err := endpoint.Serve()
	if err != nil {
		if err == io.EOF {
			log.Printf("Client disconnect: %v", conn.RemoteAddr())
		} else {
			log.Printf("Error on %v", err)
		}
	}

	s.lock.Lock()
	log.Printf("Deleting worker %v", conn.RemoteAddr())
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
	log.Printf("connecting to #%d, %s ...\n", order.Id, order.Address())

	errch := make(chan error, 1)
	pool, err := NewPool(order, errch)
	if err != nil {
		log.Printf("Failed to connecting the pool %s: %s\n", order.Address(), err.Error())
		order.markDead()
		return
	}

	s.ActivePool(order, pool, errch)
}

func (s *StratumServer) ActivePool(order *Order, pool *Pool, errch chan error) {
	s.lock.Lock()
	s.perrchs[order.Id] = errch
	s.pools[order.Id] = pool
	s.lock.Unlock()
}

func (s *StratumServer) Shutdown() {
	s.stopListen()
	s.stopPools()
	s.stopWorkers()
}

func (s *StratumServer) stopListen() {
	s.closing = true
}

func (s *StratumServer) stopPools() {
	for _, pool := range s.pools {
		pool.Shutdown()
		// TODO: remove it from pools list?
	}
}

func (s *StratumServer) stopWorkers() {
	log.Printf("Stopping %d workers.", len(s.workers))
	for _, worker := range s.workers {
		worker.Close() // close codec
	}
}

func (s *StratumServer) firstPool() (*Pool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, pool := range s.pools {
		return pool, nil
	}

	return nil, errors.New("No pool available.")
}

// func (s *StratumServer) Connection(e *birpc.Endpoint) (conn *Connection, err error) {
// 	conn, ok := s.connections[e]
// 	if !ok {
// 		e.Close()
// 		return nil, ErrServerUnexpected
// 	}

// 	return conn, nil
// }

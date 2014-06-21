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
	workers map[*birpc.Endpoint]*birpc.Endpoint
	pools   map[uint64]*Pool
	perrchs map[uint64]chan error // pool error chans
	orders  map[uint64]*Order
	errCh   chan error
	sigCh   chan os.Signal
	closing bool
}

func NewStratumServer() *StratumServer {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	DefaultServer = &StratumServer{
		Stratum: s,
		workers: make(map[*birpc.Endpoint]*birpc.Endpoint),
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

		go func() {
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
		}()
	}
}

func (s *StratumServer) newEndpoint(conn net.Conn) *birpc.Endpoint {
	e := birpc.NewEndpoint(jsonmsg.NewCodec(conn), s.registry)
	s.lock.Lock()
	s.workers[e] = e
	s.lock.Unlock()
	return e
}

func (s *StratumServer) startPools() {
	for oid, order := range s.orders {
		// test if actived
		_, ok := s.pools[oid]
		if ok {
			continue
		}

		// connect to upstream pool
		log.Printf("connecting to #%d, %s ...\n", oid, order.Address())

		errch := make(chan error, 1)
		s.perrchs[order.Id] = errch
		pool, err := NewPool(order, errch)
		if err != nil {
			log.Printf("Failed to connecting the pool %s: %s\n", order.Address(), err.Error())
			order.markDead()
			continue
		}

		s.lock.Lock()
		s.pools[oid] = pool
		s.lock.Unlock()
	}
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

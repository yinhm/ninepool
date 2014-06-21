package stratum

import (
	"errors"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/birpc/jsonmsg"
	"io"
	"log"
	"net"
	"sync"
)

var ErrServerUnexpected = errors.New("Server error.")
var DefaultServer *StratumServer

func NewServer(ln net.Listener) {
	s := NewStratumServer()
	defer s.close()

	go s.startPools()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func() {
			s.Serve(conn)
		}()
	}
}

type StratumServer struct {
	lock sync.Mutex
	*Stratum
	endpoints   map[*birpc.Endpoint]*birpc.Endpoint
	pools       map[uint64]*Pool
	perrchs     map[uint64]chan error // pool error chans
	orders      map[uint64]*Order
}

func NewStratumServer() *StratumServer {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	DefaultServer = &StratumServer{
		Stratum:     s,
		endpoints:   make(map[*birpc.Endpoint]*birpc.Endpoint),
		pools:     make(map[uint64]*Pool),
		perrchs:     make(map[uint64]chan error),
		orders:      InitOrders("x11"),
	}
	mining := &Mining{}
	// ss.registry.RegisterService(ss)
	DefaultServer.registry.RegisterService(mining)
	return DefaultServer
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
			log.Printf("Error on connecting to %s: %s\n", order.Address(), err.Error())
			order.markDead()
			continue
		}

		s.lock.Lock()
		s.pools[oid] = pool
		s.lock.Unlock()
	}
}

func (s *StratumServer) Serve(conn net.Conn) {
	defer conn.Close()

	endpoint := s.newEndpoint(conn)
	log.Printf("Client connected: %v\n", conn.RemoteAddr())
	err := endpoint.Serve()
	if err != nil {
		if err == io.EOF {
			log.Printf("Client disconnect: %v", conn.RemoteAddr())
		} else {
			log.Printf("Error %v: %v", conn.RemoteAddr(), err)
		}
	}
}

func (s *StratumServer) newEndpoint(conn net.Conn) *birpc.Endpoint {
	e := birpc.NewEndpoint(jsonmsg.NewCodec(conn), s.registry)
	s.endpoints[e] = e
	return e
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

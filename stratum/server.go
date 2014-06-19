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
var mainserver *StratumServer

func NewServer(ln net.Listener) {
	s := NewStratumServer()
	defer s.close()

	go s.warmProxies()

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
	connections map[*birpc.Endpoint]*Connection
	proxies     map[uint64]*Proxy
	orders      map[uint64]*Order
}

func NewStratumServer() *StratumServer {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	mainserver = &StratumServer{
		Stratum:     s,
		connections: make(map[*birpc.Endpoint]*Connection),
		proxies:     make(map[uint64]*Proxy),
		orders:      InitOrders("x11"),
	}
	mining := &Mining{}
	// ss.registry.RegisterService(ss)
	mainserver.registry.RegisterService(mining)
	return mainserver
}

func (s *StratumServer) warmProxies() {
	for oid, order := range s.orders {
		// test if actived
		_, ok := s.proxies[oid]
		if ok {
			continue
		}

		// connect to upstream proxy
		log.Printf("connecting to #%d, %s ...\n", oid, order.Address())

		proxy, err := NewProxy(order)
		if err != nil {
			log.Printf("Error on connecting to %s: %s\n", order.Address(), err.Error())
			order.markDead()
			continue
		}

		s.lock.Lock()
		s.proxies[oid] = proxy
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
	clientConn := &Connection{endpoint: e}
	// clientConn.bindProxy(s.firstOrder())
	s.connections[e] = clientConn
	return e
}

func (s *StratumServer) firstProxy() (*Proxy, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, proxy := range s.proxies {
		return proxy, nil
	}

	return nil, errors.New("No proxy available.")
}

func (s *StratumServer) Connection(e *birpc.Endpoint) (conn *Connection, err error) {
	conn, ok := s.connections[e]
	if !ok {
		e.Close()
		return nil, ErrServerUnexpected
	}

	return conn, nil
}

type Proxy struct {
	address  string
	order    *Order
	upstream *StratumClient
	miners   map[*birpc.Endpoint]*Connection
	closing  bool
}

func NewProxy(order *Order) (proxy *Proxy, err error) {
	conn, err := net.Dial("tcp", order.Address())
	if err != nil {
		return nil, err
	}

	upstream := NewClient(conn)
	err = upstream.Subscribe()
	if err != nil {
		return nil, err
	}

	err = upstream.Authorize("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "x")
	if err != nil {
		return nil, err
	}

	order.markConnected()

	p := &Proxy{
		address:  order.Address(),
		order:    order,
		upstream: upstream,
	}

	return p, nil
}

package stratum

import (
	"errors"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/birpc/jsonmsg"
	"io"
	"log"
	"net"
)

var ErrServerUnexpected = errors.New("Server error.")
var mainserver *StratumServer

func NewServer(ln net.Listener) {
	s := NewStratumServer()
	defer s.close()

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
	*Stratum
	connections map[*birpc.Endpoint]*Connection
	orders map[uint64]*Order
}

func NewStratumServer() *StratumServer {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	mainserver = &StratumServer{
		Stratum:     s,
		connections: make(map[*birpc.Endpoint]*Connection),
	}
	mining := &Mining{}
	// ss.registry.RegisterService(ss)
	mainserver.registry.RegisterService(mining)
	return mainserver
}

func (s *StratumServer) newEndpoint(conn net.Conn) *birpc.Endpoint {
	e := birpc.NewEndpoint(jsonmsg.NewCodec(conn), s.registry)
	s.connections[e] = &Connection{endpoint: e}
	return e
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
	miners   *map[*birpc.Endpoint]*Connection
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

	p := &Proxy{
		address:  order.Address(),
		order:    order,
		upstream: upstream,
	}

	return p, nil
}

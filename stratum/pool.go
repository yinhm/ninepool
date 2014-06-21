package stratum

import (
	"github.com/yinhm/ninepool/birpc"
	"net"
)

type Pool struct {
	address  string
	order    *Order
	upstream *StratumClient
	miners   map[*birpc.Endpoint]*birpc.Endpoint
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

	p := &Pool{
		address:  order.Address(),
		order:    order,
		upstream: upstream,
	}

	return p, nil
}

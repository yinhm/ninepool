package stratum

import (
	"fmt"
	"github.com/tv42/birpc"
	"github.com/tv42/birpc/jsonmsg"
	"github.com/tv42/topic"
	"io"
	"net"
)

func NewClient(conn net.Conn) *StratumClient {
	c := NewStratumClient()
	defer c.close()

	c.Serve(conn)
	return c
}

type StratumClient struct {
	*Stratum
	endpoint *birpc.Endpoint
}

func NewStratumClient() *StratumClient {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	sc := &StratumClient{
		Stratum: s,
	}
	mining := &Mining{}
	// sc.registry.RegisterService(sc)
	sc.registry.RegisterService(mining)
	return sc
}

func (c *StratumClient) Serve(conn io.ReadWriteCloser) {
	c.endpoint = birpc.NewEndpoint(jsonmsg.NewCodec(conn), c.registry)
	errCh := make(chan error)
	go func() {
		errCh <- c.endpoint.Serve()
	}()
}

func (c *StratumClient) Subscribe() {
	args := []interface{}{}
	reply := &[]interface{}{}
	err := c.endpoint.Call("mining.subscribe", args, reply)

	fmt.Printf("%v\n", reply)

	if err != nil {
		fmt.Printf("unexpected error from call: %v\n", err.Error())
	}
}

func (c *StratumClient) Authorize(username, password string) {
	var authed bool
	params := []interface{}{username, password}
	err := c.endpoint.Call("mining.authorize", params, &authed)
	if err != nil {
		fmt.Printf("not expected: %v\n", err.Error())
	}
	fmt.Printf("auth res: %v\n", authed)
}

func (c *StratumClient) Submit(username, jobId, extranonce2, ntime, nonce string) {
	var accepted bool
	params := []interface{}{username, jobId, extranonce2, ntime, nonce}
	err := c.endpoint.Call("mining.submit", params, &accepted)
	if err != nil {
		fmt.Printf("share rejected: %v\n", err.Error())
	}
	fmt.Printf("share accepted: %v\n", accepted)
}

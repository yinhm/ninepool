package stratum

import (
	"log"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/birpc/jsonmsg"
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
	authorized      bool
	extraNonce1     string
	extraNonce2Size uint64
	prevDifficulty  uint64
	difficulty      uint64
	remoteAddress   string
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

func (c *StratumClient) Subscribe() (err error) {
	args := []interface{}{}
	reply := &[]interface{}{}
	err = c.endpoint.Call("mining.subscribe", args, reply)

	if err != nil {
		log.Printf("mining.subscribe failed: %v\n", err.Error())
		return err
	}

	log.Printf("%v", reply)
	return nil
}

func (c *StratumClient) Authorize(username, password string) {
	var authed bool
	params := []interface{}{username, password}
	err := c.endpoint.Call("mining.authorize", params, &authed)
	if err != nil {
		log.Printf("not expected: %v\n", err.Error())
	}
	log.Printf("auth res: %v\n", authed)
}

func (c *StratumClient) Submit(username, jobId, extranonce2, ntime, nonce string) {
	var accepted bool
	params := []interface{}{username, jobId, extranonce2, ntime, nonce}
	err := c.endpoint.Call("mining.submit", params, &accepted)
	if err != nil {
		log.Printf("share rejected: %v\n", err.Error())
	}
	log.Printf("share accepted: %v\n", accepted)
}

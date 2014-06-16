package stratum

import (
	"errors"
	"fmt"
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
	active          bool
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
	args := List{}
	reply := &List{}
	err = c.endpoint.Call("mining.subscribe", args, reply)

	if err != nil {
		return errors.New("mining.subscribe failed")
	}

	data := (List)(*reply)

	c.extraNonce1 = data[1].(string)
	if c.extraNonce1 == "" {
		return errors.New("Failed to get nonce1")
	}

	c.extraNonce2Size = (uint64)(data[2].(float64))
	if c.extraNonce2Size < 1 {
		return errors.New("Failed to get nonce2size")
	}

	c.active = true

	return nil
}

func (c *StratumClient) Authorize(username, password string) error {
	var authed bool
	params := List{username, password}
	err := c.endpoint.Call("mining.authorize", params, &authed)
	if err != nil {
		return errors.New("Auth failed.")
	}

	c.authorized = true
	return nil
}

func (c *StratumClient) Submit(username, jobId, extranonce2, ntime, nonce string) error {
	var accepted bool
	params := List{username, jobId, extranonce2, ntime, nonce}
	err := c.endpoint.Call("mining.submit", params, &accepted)
	if err != nil {
		return errors.New(fmt.Sprintf("share rejected, %s.", err.Error()))
	}

	log.Printf("share accepted: %v\n", accepted)
	return nil
}

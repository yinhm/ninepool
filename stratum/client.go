package stratum

import (
	"errors"
	"fmt"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/birpc/jsonmsg"
	"io"
	"log"
	"net"
)

func NewClient(conn net.Conn, errch chan error) *StratumClient {
	c := NewStratumClient()
	defer c.close()

	c.Serve(conn, errch)
	return c
}

// Stratum client context, passed to birpc
type ClientContext struct {
	SubId           string
	OrderId         uint64
	Authorized      bool
	ExtraNonce1     string
	ExtraNonce2Size uint64
	PrevDifficulty  float64
	Difficulty      float64
	RemoteAddress   string
}

type StratumClient struct {
	*Stratum
	endpoint *birpc.Endpoint
	context  *ClientContext
	active   bool
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

func (c *StratumClient) Serve(conn io.ReadWriteCloser, errch chan error) {
	c.endpoint = birpc.NewEndpoint(jsonmsg.NewCodec(conn), c.registry)
	c.endpoint.Context = &ClientContext{}
	go func() {
		err := c.endpoint.Serve()
		if err != nil {
			errch <- err
		}
	}()
}

func (c *StratumClient) Close() {
	c.endpoint.Close()
}

func (c *StratumClient) Subscribe() (err error) {
	args := birpc.List{}
	reply := &birpc.List{}
	err = c.endpoint.Call("mining.subscribe", args, reply)

	if err != nil {
		return errors.New("mining.subscribe failed")
	}

	data := (birpc.List)(*reply)

	context := c.endpoint.Context.(*ClientContext)
	context.ExtraNonce1 = data[1].(string)
	if context.ExtraNonce1 == "" {
		return errors.New("Failed to get nonce1")
	}

	context.ExtraNonce2Size = (uint64)(data[2].(float64))
	if context.ExtraNonce2Size < 1 {
		return errors.New("Failed to get nonce2size")
	}

	c.active = true

	return nil
}

func (c *StratumClient) Authorize(username, password string) error {
	var authed bool
	params := birpc.List{username, password}
	err := c.endpoint.Call("mining.authorize", params, &authed)
	if err != nil {
		return errors.New("Auth failed.")
	}

	context := c.endpoint.Context.(*ClientContext)
	context.Authorized = true
	return nil
}

func (c *StratumClient) Submit(username, jobId, extranonce2, ntime, nonce string) error {
	var accepted bool
	params := birpc.List{username, jobId, extranonce2, ntime, nonce}
	err := c.endpoint.Call("mining.submit", params, &accepted)
	if err != nil {
		return errors.New(fmt.Sprintf("share rejected, %s.", err.Error()))
	}

	log.Printf("share accepted: %v\n", accepted)
	return nil
}

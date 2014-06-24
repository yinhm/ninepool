package stratum_test

import (
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/stratum"
	"io"
	"net"
	"testing"
	"time"
)

const (
	FOOBAR           = `FOOBAR\n`
	MINING_SUBSCRIBE = `{"id":1,"method":"mining.subscribe","params":[]}` + "\n"
)

var cli, srv net.Conn
var server *stratum.StratumServer

func initServer() {
	cli, srv = net.Pipe()

	options := stratum.Options{
		SubscribeTimeout: time.Duration(100) * time.Millisecond,
	}
	server = stratum.NewStratumServer(options)
	go server.ServeConn(srv)
}

func addOrder() {
	order := &stratum.Order{
		Id:       1,
		Hostname: "112.124.104.176",
		Port:     "3333",
		Username: "1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda",
		Password: "x",
	}
	server.AddOrder(order)

	// active mock order
	pcli, _ := net.Pipe()
	errch := make(chan error, 1)
	upstream := stratum.NewClient(pcli, errch)
	ctx := upstream.Context()
	ctx.ExtraNonce2Size = 4
  ctx.CurrentJob = &stratum.Job{
    "bf",
    "4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000",
    "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
    "072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
		birpc.List{},
    "00000002",
    "1c2ac4af",
    "504e86b9",
    false,
  }

	p, _ := stratum.NewPoolWithConn(order, upstream)
	server.ActivePool(order, p, errch)
	// _ = p.Context()
}

func closeServer() {
	server.Shutdown()
	cli.Close()
	srv.Close()
}

func TestSubscribe(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	if client.Active != true {
		t.Fatalf("Client not active.")
	}

	closeServer()
}

func TestSubscribeTimeout(t *testing.T) {
	initServer()
	addOrder()

	time.Sleep(150 * time.Millisecond)

	_, err := io.WriteString(cli, MINING_SUBSCRIBE)
	if err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatalf("client should timeout without subscribe: %v", err)
	}

	closeServer()
}

func TestSubscribeNoPool(t *testing.T) {
	initServer()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err == nil || err.Error() != "No pool available" {
		t.Fatalf("Should not have pools available: %v", err)
	}

	closeServer()
}

func TestSetDifficulty(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	if client.Active != true {
		t.Fatalf("Client not active.")
	}

	time.Sleep(20 * time.Millisecond) // wait for notification
	ctx := client.Context()
	if ctx.Difficulty != stratum.DefaultDifficulty {
		t.Fatalf("mining.set_difficulty not received.")
	}
	if ctx.CurrentJob.JobId != "bf" {
		t.Fatalf("mining.notify not received.")
	}

	closeServer()
}

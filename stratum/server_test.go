package stratum_test

import (
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
	p := stratum.NewPoolWithConn(order, upstream)

	server.ActivePool(order, p, errch)
}

func closeServer() {
	server.Shutdown()
	cli.Close()
	srv.Close()
}

func TestSubscribe(t *testing.T) {
	initServer()
	defer closeServer()
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
}

func TestSubscribeTimeout(t *testing.T) {
	initServer()
	defer closeServer()
	server.AddOrder(&stratum.Order{Id: 1})

	time.Sleep(150 * time.Millisecond)

	_, err := io.WriteString(cli, MINING_SUBSCRIBE)
	if err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatalf("client should timeout without subscribe: %v", err)
	}
}

func TestSubscribeNoPool(t *testing.T) {
	initServer()
	defer closeServer()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err == nil || err.Error() != "No pool available" {
		t.Fatalf("Should not have pools available: %v", err)
	}
}

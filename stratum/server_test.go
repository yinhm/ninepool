package stratum_test

import (
	"github.com/yinhm/ninepool/stratum"
	"testing"
	"time"
	"net"
	"io"
)

const (
	FOOBAR = `FOOBAR\n`
	MINING_SUBSCRIBE = `{"id":1,"method":"mining.subscribe","params":[]}` + "\n"
)

func TestSubscribe(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()

	options := stratum.Options{
		SubscribeTimeout: time.Duration(200)*time.Millisecond,
	}
	service := stratum.NewStratumServer(options)
	service.AddOrder(&stratum.Order{Id: 1})
	go service.ServeConn(srv)
	
	_, err := io.WriteString(cli, MINING_SUBSCRIBE)
	if err != nil {
		t.Fatalf("client subscribed failed: %v", err)
	}
}

func TestSubscribeTimeout(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()

	options := stratum.Options{
		SubscribeTimeout: time.Duration(50)*time.Millisecond,
	}
	service := stratum.NewStratumServer(options)
	service.AddOrder(&stratum.Order{Id: 1})
	go service.ServeConn(srv)
	
	time.Sleep(100 * time.Millisecond)

	_, err := io.WriteString(cli, MINING_SUBSCRIBE)
	if err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatalf("client should timeout without subscribe: %v", err)
	}
}

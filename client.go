package main

import (
	"github.com/golang/glog"
	"github.com/yinhm/ninepool/stratum"
	"net"
	"time"
)

const (
	network = "tcp4"
	// addr    = "112.124.104.176:18333"
	// addr    = "stratum.nicehash.com:3333"
	addr = ":3335"
)

func main() {
	conn, err := net.Dial(network, addr)
	if err != nil {
		panic(err)
	}

	errch := make(chan error)
	client := stratum.NewClient(conn, errch)
	glog.Infof("client started...\n")

	go func() {
		if err := <-errch; err != nil {
			glog.Fatalf("Error on client.Serve: %s", err.Error())
		}
	}()

	err = client.Subscribe()
	if err != nil {
		glog.Fatalf(err.Error())
	}

	err = client.Authorize("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "x")
	if err != nil {
		glog.Fatalf(err.Error())
	}

	time.Sleep(200 * time.Millisecond)

	err = client.Submit("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "bf",
		"00000001", "504e86ed", "b2957c02")
	if err != nil {
		glog.Fatalf(err.Error())
	}

	time.Sleep(500 * time.Millisecond)
}

package main

import (
	"github.com/golang/glog"
	"github.com/yinhm/ninepool/stratum"
	"net"
)

const (
	network = "tcp4"
	addr    = ":3335"
)

func main() {
	options, err := stratum.ParseCommandLine()
	if err != nil {
		glog.Infof("Failed to parse command line: %s\n", err)
		return
	}

	ln, err := net.Listen(network, addr)
	if err != nil {
		glog.Fatalf("Can not bind : %s\n", err)
	}
	defer ln.Close()

	glog.Infof("Listen on %s", addr)

	service := stratum.NewStratumServer(options)
	if err := service.Start(ln); err != nil {
		glog.Fatalf("Service exited with error: %s\n", err)
	} else {
		glog.Infof("Service exited.")
	}
}

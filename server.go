package main

import (
	"github.com/yinhm/ninepool/stratum"
	"net"
)

const (
	network = "tcp4"
	addr    = ":3335"
)

func main() {
	ln, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}

	stratum.NewServer(ln)
}

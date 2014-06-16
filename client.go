package main

import (
	"fmt"
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

	client := stratum.NewClient(conn)
	fmt.Printf("client started...\n")

	client.Subscribe()
	client.Authorize("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "x")

	time.Sleep(200 * time.Millisecond)

	client.Submit("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "bf",
		"00000001", "504e86ed", "b2957c02")

	time.Sleep(500 * time.Millisecond)
}

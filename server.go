package main

import (
	"github.com/yinhm/ninepool/stratum"
	"log"
	"net"
	"os"
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
	defer ln.Close()

	log.Printf("Listen on %s", addr)

	service := stratum.NewStratumServer()
	if err := service.Start(ln); err != nil {
		log.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	} else {
		log.Println("Service exited.")
	}
}

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
	options, err := stratum.ParseCommandLine()
	if err != nil {
		log.Printf("Failed to parse command line: %s\n", err)
		return
	}

	ln, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	log.Printf("Listen on %s", addr)

	service := stratum.NewStratumServer(options)
	if err := service.Start(ln); err != nil {
		log.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	} else {
		log.Println("Service exited.")
	}
}

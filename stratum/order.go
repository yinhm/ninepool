package stratum

import (
	"fmt"
)

// Order state codes.
const (
	StateInit = iota
	StateConnected
	StateBanned
	StateWorking
	StatePause
	StateClosedCannel
	StateClosedComplete
)

type Order struct {
	Id       uint64

	Algorithm string
	// in satoshi, 10**8 staoshi = 1 btc
	Amount uint64
	Price uint64

	// Pool detail
	Hostname string
	Port     string
	Username string
	Password string

	Status   string
	Created  uint64
}

func InitOrders(algo string) []Order {
	
}

func (od *Order) Address() string {
	return fmt.Sprintf("%s:%s", od.Hostname, od.Port)
}

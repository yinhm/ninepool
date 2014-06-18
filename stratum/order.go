package stratum

import (
	"fmt"
	"math"
	"time"
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

	STAOSHI = 1 << (10 * iota)
)

var (
	UNIT_SATOSHI = uint64(math.Pow(float64(10), float64(8)))
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

	Status   uint32
	Created  uint64
}

func InitOrders(algo string) []Order {
	o = &Order{
		Id: 1,
		Algorithm: 'x11',
		Amount: 1 * UNIT_SATOSHI,
		Price: 0.05 * UNIT_SATOSHI,
		Hostname: "112.124.104.176",
		Port: "18333",
		Username: "1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda",
		Password: "x",
		Status: StateInit,
		Created: time.Now().Unix(),
	}
}

func (od *Order) Address() string {
	return fmt.Sprintf("%s:%s", od.Hostname, od.Port)
}

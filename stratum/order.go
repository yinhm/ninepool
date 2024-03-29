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
	StateDead
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
	Id uint64

	Algorithm string
	// in satoshi, 10**8 staoshi = 1 btc
	Amount uint64
	Price  uint64

	// Pool detail
	Hostname string
	Port     string
	Username string
	Password string

	State   uint32
	Created int64
}

func InitOrders(algo string) map[uint64]*Order {
	od := &Order{
		Id:        1,
		Algorithm: algo,
		Amount:    1 * UNIT_SATOSHI,
		Price:     5 * UNIT_SATOSHI / 100,
		Hostname:  "localhost",
		Port:      "3334",
		Username:  "n1jBXLw6eeFSdp1sznxvSLVgjJP4Ag7bRh",
		Password:  "x",
		State:     StateInit,
		Created:   time.Now().Unix(),
	}

	orders := make(map[uint64]*Order)
	orders[od.Id] = od
	return orders

}

func (od *Order) Address() string {
	return fmt.Sprintf("%s:%s", od.Hostname, od.Port)
}

func (od *Order) markDead() {
	od.State = StateDead
}

func (od *Order) markConnected() {
	od.State = StateConnected
}

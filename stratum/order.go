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
	Id       string
	Hostname string
	Port     string
	Username string
	Password string
	Status   string
	Created  int64
}

func (od *Order) Address() string {
	return fmt.Sprintf("%s:%s", od.Hostname, od.Port)
}

package stratum

import (
	"container/ring"
	"errors"
	// "github.com/golang/glog"
	"time"
)

type Share struct {
	username string
	// jobId    string
	pool     string //hash?
	header string
	diff float64
	isBlock  bool
	ntime  float64
}

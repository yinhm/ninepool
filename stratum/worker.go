package stratum

import (
	"github.com/yinhm/ninepool/birpc"
	"log"
	"sync"
	"time"
)

// Stratum connection context, passed to birpc
type Context struct {
	SubId           string
	OrderId         uint64
	Authorized      bool
	ExtraNonce1     string
	ExtraNonce2Size uint64
	PrevDifficulty  float64
	Difficulty      float64
	RemoteAddress   string
	SubCh           chan bool
}

type Worker struct {
	lock         sync.Mutex
	endpoint     *birpc.Endpoint
	context      *Context
	pool         *Pool
	connected    bool
	samplePeriod int // in minutes
	accepted     int
	rejected     int
	created      int64
}

func NewWorker(pool *Pool, endpoint *birpc.Endpoint, timeout time.Duration) *Worker {
	context := &Context{
		SubCh: make(chan bool, 1),
	}

	endpoint.Context = context

	worker := &Worker{
		endpoint:     endpoint,
		context:      context,
		pool:         pool,
		samplePeriod: 600,
		created:      time.Now().Unix(),
	}

	go worker.ensureSubscribe(timeout)

	return worker
}

func (w *Worker) Close() {
	w.endpoint.Close()
}

func (w *Worker) ensureSubscribe(timeout time.Duration) {
	//go func() { c <- client.Call("Service.Method", args, &reply) } ()
	select {
	case _ = <-w.context.SubCh:
		w.connected = true
		return
	case <-time.After(timeout):
		log.Printf("No subscribe request received from worker in %d seconds.", timeout)
		w.Close()
		return
	}
}

func (w *Worker) rebind(newPool *Pool) {
	w.pool = newPool
}

func (w *Worker) newExtraNonce() {

}

func (w *Worker) newDifficulty() {

}

// Update the shares lists with the given share to compute hashrate
func (w *Worker) updateShareLists() {

}

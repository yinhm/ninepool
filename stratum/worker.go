package stratum

import (
	"errors"
	"github.com/yinhm/ninepool/birpc"
	"log"
	"sync"
	"time"
)

type Worker struct {
	lock         sync.Mutex
	endpoint     *birpc.Endpoint
	context      *Context
	connected    bool // true when subscribed
	samplePeriod int  // in minutes
	accepted     int
	rejected     int
	created      int64
	closing      bool
}

func NewWorker(endpoint *birpc.Endpoint, timeout time.Duration) *Worker {
	context := &Context{
		SubCh:  make(chan bool, 1),
		PoolCh: make(chan bool, 1),
	}

	endpoint.Context = context

	worker := &Worker{
		endpoint:     endpoint,
		context:      context,
		samplePeriod: 600,
		created:      time.Now().Unix(),
	}

	go worker.waitSubscribe(timeout)

	return worker
}

func (w *Worker) Close() {
	w.lock.Lock()
	w.lock.Unlock()

	if w.closing == true {
		return
	}
	w.detachPool()
	w.endpoint.Close()
	w.closing = true
}

func (w *Worker) waitSubscribe(timeout time.Duration) {
	select {
	case _ = <-w.context.SubCh:
		err := w.bindFirstPool()
		if err != nil {
			w.context.PoolCh <- false
			w.connected = false
		} else {
			w.context.PoolCh <- true
			w.connected = true
		}
		return
	case <-time.After(timeout):
		log.Printf("No subscribe request received from worker in %.2f seconds.", timeout.Seconds())
		w.Close()
		return
	}
}

func (w *Worker) bindFirstPool() error {
	pool, err := DefaultServer.firstPool()
	if err != nil {
		return err
	}
	w.rebind(pool)
	return nil
}

func (w *Worker) rebind(newPool *Pool) {
	w.detachPool()
	w.context.pool = newPool
	w.context.pool.addWorker(w)
}

func (w *Worker) detachPool() {
	if w.context.pool == nil {
		return
	}
	w.context.pool.removeWorker(w)
	w.context.pool = nil
}

func (w *Worker) sendJob(job *Job) {
	var msg birpc.Message
	msg.ID = 0
	msg.Func = "mining.notify"
	msg.Args = job.tolist()
	w.endpoint.Notify(&msg)
}

func (w *Worker) newExtraNonce() {

}

func (w *Worker) newDifficulty() {

}

// Update the shares lists with the given share to compute hashrate
func (w *Worker) updateShareLists() {

}

// Stratum connection context, passed to birpc
type Context struct {
	pool            *Pool
	SubId           string
	Username        string
	Password        string
	OrderId         uint64
	Authorized      bool
	ExtraNonce1     string
	ExtraNonce2Size int
	PrevDifficulty  float64
	Difficulty      float64
	RemoteAddress   string
	SubCh           chan bool
	PoolCh          chan bool // pool available
}

func (ctx *Context) CurrentJob() (*Job, error) {
	if ctx.pool == nil {
		return nil, errors.New("no pool avilable")
	}
	return ctx.pool.CurrentJob, nil
}

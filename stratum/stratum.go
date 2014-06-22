package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"log"
	"sync"
	"time"
	"unsafe"
)

type Stratum struct {
	broadcast *topic.Topic
	registry  *birpc.Registry
}

func NewStratum() *Stratum {
	s := &Stratum{
		broadcast: topic.New(),
		registry:  birpc.NewRegistry(),
	}
	s.registry.RegisterService(s)

	return s
}

func (s *Stratum) close() {
	close(s.broadcast.Broadcast)
}

type Mining struct{}

func (m *Mining) Subscribe(req *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	context := e.Context.(*Context)

	context.SubCh <- true
	succeed := <-context.PoolCh
	if !succeed {
		e.WaitClose()
		return &birpc.Error{ErrorUnknown, "No pool available", nil}
	}

	subId := randhash()
	context.SubId = subId

	*reply = birpc.List{
		[][]string{
			{"mining.set_difficulty", subId},
			{"mining.notify", subId},
		},
		"08000002",
		4,
	}

	defer m.notify(e)
	return nil
}

func (m *Mining) notify(e *birpc.Endpoint) {
	time.Sleep(50 * time.Millisecond)

	newjob := []interface{}{
		"bf",
		"4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
		"00000002",
		"1c2ac4af",
		"504e86b9",
		false,
	}

	var msg birpc.Message

	msg.ID = 0
	msg.Func = "mining.notify"
	msg.Args = nil
	msg.Result = newjob

	e.Notify(&msg)
}

// mining.notify notification from upstream
func (m *Mining) Notify(req *interface{}, reply *interface{}) error {
	log.Printf("mining.Notify\n")
	return nil
}

// mining.set_difficulty notification from upstream
func (m *Mining) Set_difficulty(req *interface{}, reply *interface{}) error {
	log.Printf("mining.set_difficulty\n")
	return nil
}

func (m *Mining) Authorize(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	username := (*args).([]interface{})[0].(string)

	_, err := btcutil.DecodeAddress(username, &btcnet.MainNetParams)
	if err != nil {
		*reply = false
		e.WaitClose()
		return err
	}

	// authented
	context := e.Context.(*Context)
	context.Authorized = true

	*reply = true
	return nil
}

func (m *Mining) Submit(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	context := e.Context.(*Context)
	// verify authentation
	if context.Authorized != true {
		e.WaitClose()
		txt, _ := errorText[ErrorUnauthorizedWorker]
		return &birpc.Error{ErrorUnauthorizedWorker, txt, nil}
	}

	params := (*args).([]interface{})
	username := params[0].(string)
	jobId := params[1].(string)
	extranonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	if context.ExtraNonce1 == "" {
		e.WaitClose()
		txt, _ := errorText[ErrorUnsubscribedWorker]
		return &birpc.Error{ErrorUnsubscribedWorker, txt, nil}
	}

	err2 := m.processShare(username, jobId, extranonce2, ntime, nonce)
	if err2 != nil {
		return err2
	}

	*reply = true
	return nil
}

func (m *Mining) processShare(username, jobId, extranonce2, ntime, nonce string) error {
	log.Printf("share accepted: %v\n", jobId)
	return nil
}

type ExtraNonceCounter struct {
	lock             sync.Mutex
	count            uint32
	Size             int
	NoncePlaceHolder []byte
}

func NewExtraNonceCounter() *ExtraNonceCounter {
	var count uint32 = 1 << 27
	p, _ := hex.DecodeString("f000000ff111111f")

	ct := &ExtraNonceCounter{
		count:            count,
		Size:             int(unsafe.Sizeof(count)),
		NoncePlaceHolder: p,
	}
	return ct
}

func (ct *ExtraNonceCounter) Next() string {
	ct.lock.Lock()
	ct.count += 1
	buf := make([]byte, ct.Size)
	binary.BigEndian.PutUint32(buf, ct.count)
	ct.lock.Unlock()
	return hex.EncodeToString(buf)
}

func (ct *ExtraNonceCounter) Nonce2Size() int {
	return len(ct.NoncePlaceHolder) - ct.Size
}

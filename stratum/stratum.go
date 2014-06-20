package stratum

import (
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"log"
	"time"
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
	*reply = birpc.List{
		[][]string{
			{"mining.set_difficulty", "b4b6693b72a50c7116db18d6497cac52"},
			{"mining.notify", "ae6812eb4cd7735a302a8a9dd95cf71f"},
		},
		"08000002",
		4,
	}

	go m.notify(e)
	return nil
}

func (m *Mining) notify(e *birpc.Endpoint) {
	time.Sleep(100 * time.Millisecond)

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
	e.Context.Authorized = true

	*reply = true
	return nil
}

func (m *Mining) Submit(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	// verify authentation
	if e.Context.Authorized != true {
		e.WaitClose()
		return &birpc.Error{24, "unauthorized worker", nil}
	}

	params := (*args).([]interface{})
	username := params[0].(string)
	jobId := params[1].(string)
	extranonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	if e.Context.ExtraNonce1 == "" {
		e.WaitClose()
		return &birpc.Error{25, "not subscribed", nil}
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

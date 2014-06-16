package stratum

import (
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/yinhm/ninepool/birpc"
	"github.com/tv42/topic"
	"log"
	"time"
)

type Response struct{}

type SubscribeParamater struct{}

type Detail struct {
	Func  string
	Value string
}

// [[["mining.set_difficulty", "b4b6693b72a50c7116db18d6497cac52"], ["mining.notify", "ae6812eb4cd7735a302a8a9dd95cf71f"]], "08000002", 4]
type SubscriptionResponse struct {
	// 2-tuple with name of subscribed notification and subscription ID
	Details          [1][2]Detail
	Extranonce1      string
	Extranonce2_size float64
}

type AuthParamater struct {
	Username string
	Password string
}

type AuthResponse bool

// job_id - ID of the job. Use this ID while submitting share generated from this job.
// prevhash - Hash of previous block.
// coinb1 - Initial part of coinbase transaction.
// coinb2 - Final part of coinbase transaction.
// merkle_branch - List of hashes, will be used for calculation of merkle root. This is not a list of all transactions, it only contains prepared hashes of steps of merkle tree algorithm. Please read some materials for understanding how merkle trees calculation works. Unfortunately this example don't have any step hashes included, my bad!
// version - Bitcoin block version.
// nbits - Encoded current network difficulty
// ntime - Current ntime/
// clean_jobs - When true, server indicates that submitting shares from previous jobs don't have a sense and such shares will be rejected. When this flag is set, miner should also drop all previous jobs, so job_ids can be eventually rotated.
type Job struct {
	JobId        string   `json:"job_id"`
	Prevhash     string   `json:"prev_hash"`
	Coinb1       string   `json:"coinb1"`
	Coinb2       string   `json:"coinb2"`
	MerkleBranch []string `json:"merkle_branch"`
	Version      string   `json:"version"`
	Nbits        string   `json:"nbits"`
	Ntime        string   `json:"ntime"`
	CleanJobs    bool     `json:"clean_jobs"`
}

// Stratum client connection context
type Connection struct {
	endpoint        *birpc.Endpoint
	authorized      bool
	extraNonce1     string
	extraNonce2Size uint64
	prevDifficulty  uint64
	difficulty      uint64
	remoteAddress   string
}

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

// func (s *Stratum) Message(msg *Incoming, _ *nothing, conn net.Conn) error {
// 	log.Printf("recv from %v:%#v\n", conn.RemoteAddr, msg)

// 	s.broadcast.Broadcast <- Outgoing{
// 		Time:    time.Now(),
// 		From:    msg.From,
// 		Message: msg.Message,
// 	}
// 	return nil
// }

type Mining struct {
}

func (m *Mining) Subscribe(req *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	*reply = []interface{}{
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
	msg.Error = nil
	msg.Func = "mining.notify"
	msg.Args = nil
	msg.Result = newjob

	e.Notify(&msg)
}

func (m *Mining) Notify(req *interface{}, reply *interface{}) error {
	log.Printf("mining.Notify\n")
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
	conn, err := mainserver.Connection(e)
	if err != nil {
		return err
	}
	conn.authorized = true

	*reply = true
	return nil
}

func (m *Mining) Submit(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	// verify authentation
	conn, err := mainserver.Connection(e)
	if err != nil {
		e.WaitClose()
		return err
	}
	if conn.authorized != true {
		e.WaitClose()
		return &birpc.Error{24, "unauthorized worker", nil}
	}

	params := (*args).([]interface{})
	username := params[0].(string)
	jobId := params[1].(string)
	extranonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	if conn.extraNonce1 == "" {
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

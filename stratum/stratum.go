package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"log"
	"math"
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

	nonce1 := context.pool.nextNonce1()
	nonce2Size := context.pool.nonce2Size()
	// if err != nil {
	// 	e.WaitClose()
	// 	return &birpc.Error{ErrorUnknown, err.Error(), nil}
	// }

	*reply = birpc.List{
		[][]string{
			{"mining.set_difficulty", subId},
			{"mining.notify", subId},
		},
		nonce1,
		nonce2Size,
	}

	context.ExtraNonce1 = nonce1
	context.ExtraNonce2Size = nonce2Size

	go m.notify(e)

	return nil
}

// server mining.notify -> client
func (m *Mining) notify(e *birpc.Endpoint) {
	time.Sleep(50 * time.Millisecond)

	context := e.Context.(*Context)

	job, err := context.CurrentJob()
	if err != nil {
		e.Close()
		return
	}

	var msg birpc.Message
	msg.ID = 0
	msg.Func = "mining.notify"
	msg.Args = job.tolist()

	e.Notify(&msg)
}

// upstream mining.notify -> server
func (m *Mining) Notify(args *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	log.Printf("mining.notify\n")

	params := birpc.List((*args).([]interface{}))
	job := Job{
		JobId:        params[0].(string),
		PrevHash:     params[1].(string),
		Coinb1:       params[2].(string),
		Coinb2:       params[3].(string),
		MerkleBranch: birpc.List(params[4].([]interface{})),
		Version:      params[5].(string),
		Nbits:        params[6].(string),
		Ntime:        params[7].(string),
		CleanJobs:    params[8].(bool),
	}

	ctx := e.Context.(*ClientContext)
	ctx.CurrentJob = &job

	ctx.JobCh <- job

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

type NonceCounter interface {
	Next() string
	Nonce2Size() int
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

// Logic should be the same as tail_iterator in stratum-mining-proxy
//
// # Proxypool #
// ## How it works (how to proxy Stratrum) ##
// The main problem to solve is how to generate unique work for every proxypool
// client.

// The idea is to reduce the size of `extraNonce2` so that the server controls
// the first few bytes. This means that server will be able to generate a
// unique coinbase for each client, mutating the coinbase hash.

// This reduces the block search space for clients. However, the impact is
// negligible. With a 4 byte upstream `extraNonce2`, and with the proxypool
// server keeping 2 bytes for itself (clients get the other 2). This allows the
// server to have 65536 concurrent connections and clients to have a maximum
// hashrate of 2^32 x 2^16, or 256 tera hashes per second, which is more than
// enough for Scrypt based coins at the time of writing.

// Upon client share submission, the server checks that the share matches the
// required upstream difficulty and resubmits it under it's own name.

// ### Nonce ###
// `extraNonce2Size` and `extraNonce3Size` control the how the upstream's
// `extraNonce2` is split. Thus `extraNonce2Size` and `extraNonce3Size` should
// add up the to the upstream's `extraNonce2`'s size.

// Zero extranonce is reserved for getwork connections.
type ProxyExtraNonceCounter struct {
	lock        sync.Mutex
	count       uint32
	maxClients  int
	extraNonce1 string
	extra1Size  int
	extra2Size  int
	extra3Size  int
}

func NewProxyExtraNonceCounter(extraNonce1 string, extra2Size, extra3Size int) *ProxyExtraNonceCounter {
	maxClients := int(math.Pow(2, float64(extra3Size*8)))

	ct := &ProxyExtraNonceCounter{
		maxClients:  maxClients,
		extraNonce1: extraNonce1,
		extra2Size:  extra2Size,
		extra3Size:  extra3Size,
	}
	ct.extra1Size = int(unsafe.Sizeof(ct.count))
	return ct
}

// TODO: what happen if nonce1 excceed max???
func (ct *ProxyExtraNonceCounter) Next() string {
	ct.lock.Lock()
	ct.count += 1
	buf := make([]byte, ct.extra1Size)
	binary.BigEndian.PutUint32(buf, ct.count)
	ct.lock.Unlock()
	index := ct.extra1Size - ct.extra2Size
	return ct.extraNonce1 + hex.EncodeToString(buf[index:])
}

// api compatible with ExtraNonceCounter
func (ct *ProxyExtraNonceCounter) Nonce2Size() int {
	return ct.extra3Size
}

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
	JobId        string
	PrevHash     string
	Coinb1       string
	Coinb2       string
	MerkleBranch birpc.List
	Version      string
	Nbits        string
	Ntime        string
	CleanJobs    bool
}

func (job *Job) tolist() *birpc.List {
	return &birpc.List{
		job.JobId,
		job.PrevHash,
		job.Coinb1,
		job.Coinb2,
		job.MerkleBranch,
		job.Version,
		job.Nbits,
		job.Ntime,
		job.CleanJobs,
	}
}

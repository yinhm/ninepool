package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/tv42/topic"
	"github.com/yinhm/ninepool/birpc"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
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

func (m *Mining) rpcError(errCode int) *birpc.Error {
	txt, _ := errorText[errCode]
	return &birpc.Error{errCode, txt, nil}
}

func (m *Mining) rpcUnknownError(errMsg string) *birpc.Error {
	return &birpc.Error{ErrorUnknown, errMsg, nil}
}

func (m *Mining) Subscribe(req *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	context := e.Context.(*Context)

	context.SubCh <- true
	succeed := <-context.PoolCh
	if !succeed {
		e.WaitClose()
		return &birpc.Error{ErrorUnknown, "No pool available", nil}
	}

	subId := randhash() // unique across server
	context.SubId = subId

	// Move to pool.addWorker?, let us reusing nonce1.
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

	go m.notifyAfterSubscribe(e)

	return nil
}

// server mining.notify -> client
func (m *Mining) notifyAfterSubscribe(e *birpc.Endpoint) {
	time.Sleep(50 * time.Millisecond)

	context := e.Context.(*Context)

	// set difficulty
	var msg birpc.Message
	msg.ID = 0
	msg.Func = "mining.set_difficulty"
	msg.Args = &birpc.List{DefaultDifficulty}
	e.Notify(&msg)

	context.PrevDifficulty = DefaultDifficulty
	context.Difficulty = DefaultDifficulty

	job, err := context.CurrentJob()
	if err != nil {
		log.Printf("No job to do, return")
		e.Close()
		return
	}

	msg.ID = 0
	msg.Func = "mining.notify"
	msg.Args = job.tolist()

	e.Notify(&msg)
}

// upstream mining.notify -> server
func (m *Mining) Notify(args *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	log.Printf("mining.notify\n")

	params := birpc.List((*args).([]interface{}))
	params[4] = birpc.List(params[4].([]interface{})) // MerkleBranch
	job, err := NewJob(params)
	if err != nil {
		log.Printf("error in build job: %s\n", err.Error())
		return err
	}

	ctx := e.Context.(*ClientContext)
	ctx.CurrentJob = job
	ctx.JobCh <- job

	return nil
}

// mining.set_difficulty notification from upstream
func (m *Mining) Set_difficulty(args *interface{}, reply *interface{}, e *birpc.Endpoint) error {
	ctx := e.Context.(*ClientContext)
	ctx.PrevDifficulty = ctx.Difficulty

	params := birpc.List((*args).([]interface{}))
	ctx.Difficulty = params[0].(float64)
	log.Printf("mining.set_difficulty to %.3f\n", ctx.Difficulty)
	return nil
}

func (m *Mining) Authorize(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	params := (*args).([]interface{})
	username := params[0].(string)
	password := params[1].(string)

	_, err := btcutil.DecodeAddress(username, &btcnet.MainNetParams)
	if err != nil {
		e.WaitClose()
		*reply = false
	} else {
		// authented
		context := e.Context.(*Context)
		context.Username = username
		context.Password = password
		context.Authorized = true

		*reply = true
	}

	return nil
}

func (m *Mining) Submit(args *interface{}, reply *bool, e *birpc.Endpoint) error {
	params := (*args).([]interface{})
	username := params[0].(string)
	jobId := params[1].(string)
	extraNonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	context := e.Context.(*Context)

	// verify authentation
	if context.Authorized != true || username != context.Username {
		// m.ban()
		e.WaitClose()
		return m.rpcError(ErrorUnauthorizedWorker)
	}

	// check extranonce1 present
	if context.ExtraNonce1 == "" {
		// m.ban()
		e.WaitClose()
		return m.rpcError(ErrorUnsubscribedWorker)
	}

	// check extranonce2 size
	submitTime := time.Now().Unix()
	if len(extraNonce2)/2 != context.ExtraNonce2Size {
		return m.rpcUnknownError("incorrect size of extranonce2")
	}

	pool := context.pool
	job, ok := pool.findJob(jobId)
	if !ok {
		return m.rpcError(ErrorJobNotFound)
	}

	if len(ntime) != 8 {
		return m.rpcUnknownError("incorrect size of ntime")
	}

	ntimeInt, _ := HexToInt64(ntime)
	if ntimeInt > submitTime+7200 {
		return m.rpcUnknownError("ntime out of range")
	}

	if len(nonce) != 8 {
		return m.rpcUnknownError("incorrect size of nonce")
	}

	submission := context.ExtraNonce1 + extraNonce2 + ntime + nonce
	if err := job.submit(submission); err != nil {
		return m.rpcError(ErrorDuplicateShare)
	}

	// target check
	merkleRoot := job.MerkleRoot(context.ExtraNonce1, extraNonce2)
	header, err := SerializeHeader(job, merkleRoot, ntime, nonce)
	if err != nil {
		return m.rpcUnknownError("job error")
	}
	headerHash, _ := header.BlockSha()

	difficulty := context.Difficulty
	shareDiff, err := shareDifficulty(&headerHash, int64(1))
	log.Printf("share diff: %.4f/%.8f", difficulty, shareDiff)
	if err != nil || shareDiff/difficulty < 0.99 {
		log.Printf("[Proxy] share rejected: low difficulty, %s", headerHash.String())
		// DEBUG: testing submit low diff share, will remove in production
		go pool.submit(shareDiff, jobId, context.ExtraNonce1, extraNonce2, ntime, nonce, headerHash.String())
		return m.rpcError(ErrorLowDifficultyShare)
	}

	go pool.submit(shareDiff, jobId, context.ExtraNonce1, extraNonce2, ntime, nonce, headerHash.String())

	*reply = true
	log.Printf("[Proxy] share accepted: #%s, hash: %s\n", jobId, headerHash.String())
	return nil
}

func shareDifficulty(headerHash *btcwire.ShaHash, multiplier int64) (float64, error) {
	// diff1 := 0x00000000FFFF0000000000000000000000000000000000000000000000000000
	compact := uint32(0x1d00ffff)
	diff1 := CompactToBig(compact)

	headerBigInt := ShaHashToBig(headerHash)
	denominator := new(big.Int).Div(headerBigInt, big.NewInt(multiplier))

	shareDiff := new(big.Rat).SetFrac(diff1, denominator)
	outString := shareDiff.FloatString(10)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		return 0, err
	}
	return diff, nil
}

func DiffToTarget(diff int64) *big.Int {
	// diff1 := 0x00000000FFFF0000000000000000000000000000000000000000000000000000
	compact := uint32(0x1d00ffff)
	diff1 := CompactToBig(compact)
	target := new(big.Int).Div(diff1, big.NewInt(diff))
	return target
}

type NonceCounter interface {
	Next() string
	// session id
	Nonce2Size() int
	Nonce1Suffix(string) string
}

type ExtraNonceCounter struct {
	lock             sync.Mutex
	count            uint32
	Size             int
	NoncePlaceHolder []byte
}

// Last 5 most-significant bits represents instance_id
// The rest is just an iterator of jobs.
func NewExtraNonceCounter() *ExtraNonceCounter {
	var count uint32 = 1 << 27
	// the same in CoinbaseTransaction, not sure where its from.
	p, _ := hex.DecodeString("f000000ff111111f")

	ct := &ExtraNonceCounter{
		count:            count,
		Size:             int(unsafe.Sizeof(count)),
		NoncePlaceHolder: p,
	}
	return ct
}

// session id, server id take 5 bits, leave 27 bits or 2**27 for workers.
func (ct *ExtraNonceCounter) Next() string {
	ct.lock.Lock()
	buf := make([]byte, ct.Size)
	binary.BigEndian.PutUint32(buf, ct.count) // start from 0000
	ct.count += 1
	ct.lock.Unlock()
	return hex.EncodeToString(buf)
}

func (ct *ExtraNonceCounter) Nonce2Size() int {
	return len(ct.NoncePlaceHolder) - ct.Size
}

// mock func
func (ct *ExtraNonceCounter) Nonce1Suffix(nonce1 string) string {
	return ""
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
// 2 ** 16 or 65536 for workers
func (ct *ProxyExtraNonceCounter) Next() string {
	ct.lock.Lock()
	buf := make([]byte, ct.extra1Size)
	binary.BigEndian.PutUint32(buf, ct.count) // start from 0000
	ct.count += 1
	ct.lock.Unlock()
	index := ct.extra1Size - ct.extra2Size
	return ct.extraNonce1 + hex.EncodeToString(buf[index:])
}

// api compatible with ExtraNonceCounter
func (ct *ProxyExtraNonceCounter) Nonce2Size() int {
	return ct.extra3Size
}

// appended nonce1 suffix by this counter
func (ct *ProxyExtraNonceCounter) Nonce1Suffix(nonce1 string) string {
	return strings.TrimPrefix(nonce1, ct.extraNonce1)
}

// job_id - ID of the job. Use this ID while submitting share generated from this job.
// prevhash - Hash of previous block.
// coinb1 - Initial part of coinbase transaction.
// coinb2 - Final part of coinbase transaction.
// merkle_branch - List of hashes, will be used for calculation of merkle root. This is not a list of all transactions, it only contains prepared hashes of steps of merkle tree algorithm.
// version - Bitcoin block version.
// nbits - Encoded current network difficulty
// ntime - Current ntime/
// clean_jobs - When true, server indicates that submitting shares from previous jobs don't have a sense and such shares will be rejected. When this flag is set, miner should also drop all previous jobs, so job_ids can be eventually rotated.
type Job struct {
	JobId        string
	PrevHash     string
	Coinb1       string
	Coinb2       string
	MerkleBranch []*btcwire.ShaHash
	Version      string
	Nbits        string
	Ntime        string
	CleanJobs    bool

	lock   sync.Mutex
	shares map[string]bool
}

func NewJob(list birpc.List) (*Job, error) {
	merkleBranches, err := MerkleHashesFromList(list[4])
	if err != nil {
		return nil, err
	}

	job := &Job{
		JobId:        list[0].(string),
		PrevHash:     list[1].(string),
		Coinb1:       list[2].(string),
		Coinb2:       list[3].(string),
		MerkleBranch: merkleBranches,
		Version:      list[5].(string),
		Nbits:        list[6].(string),
		Ntime:        list[7].(string),
		CleanJobs:    list[8].(bool),
		shares:       make(map[string]bool),
	}
	return job, nil
}

func MerkleHashesFromList(list interface{}) ([]*btcwire.ShaHash, error) {
	hashList := list.(birpc.List)
	merkleBranches := make([]*btcwire.ShaHash, len(hashList))
	for i, h := range hashList {
		txHash, err := NewShaHashFromMerkleBranch(h.(string))
		if err != nil {
			return nil, err
		}
		merkleBranches[i] = txHash
	}
	return merkleBranches, nil
}

func (job *Job) tolist() *birpc.List {
	merkleHashes := make([]string, len(job.MerkleBranch))
	for i, h := range job.MerkleBranch {
		merkleHashes[i] = HexToString(h.Bytes())
	}

	return &birpc.List{
		job.JobId,
		job.PrevHash,
		job.Coinb1,
		job.Coinb2,
		merkleHashes,
		job.Version,
		job.Nbits,
		job.Ntime,
		job.CleanJobs,
	}
}

func (job *Job) submit(share string) error {
	if _, ok := job.shares[share]; ok {
		return errors.New("duplicated share")
	}
	job.lock.Lock()
	job.shares[share] = true
	job.lock.Unlock()
	return nil
}

func (job *Job) buildCoinbase(nonce1, nonce2 string) []byte {
	return CoinbaseHash(job.Coinb1, nonce1, nonce2, job.Coinb2)
}

func (job *Job) MerkleRoot(nonce1, nonce2 string) *btcwire.ShaHash {
	hashes := make([]*btcwire.ShaHash, len(job.MerkleBranch)+1)
	copy(hashes[1:], job.MerkleBranch)
	coinbase := job.buildCoinbase(nonce1, nonce2)
	coinbaseHash, _ := btcwire.NewShaHash(coinbase)
	hashes[0] = coinbaseHash
	return MerkleRootFromBranches(hashes)
}

func MerkleRootFromBranches(branches []*btcwire.ShaHash) *btcwire.ShaHash {
	root := branches[0]
	if len(branches) == 1 {
		return root
	}
	for _, node := range branches[1:] {
		root = HashMerkleBranches(root, node)
	}
	return root
}

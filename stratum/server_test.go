package stratum_test

import (
	"github.com/yinhm/birpc"
	"github.com/yinhm/ninepool/stratum"
	"io"
	"net"
	"testing"
	"time"
)

const (
	FOOBAR           = `FOOBAR\n`
	MINING_SUBSCRIBE = `{"id":1,"method":"mining.subscribe","params":[]}` + "\n"
)

var cli, srv net.Conn
var server *stratum.StratumServer

func initServer() {
	cli, srv = net.Pipe()

	options := stratum.Options{
		SubscribeTimeout: time.Duration(100) * time.Millisecond,
	}
	server = stratum.NewStratumServer(options)
	go server.ServeConn(srv)
}

func addOrder() {
	order := &stratum.Order{
		Id:       1,
		Hostname: "112.124.104.176",
		Port:     "3333",
		Username: "1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda",
		Password: "x",
	}
	server.AddOrder(order)

	// active mock order
	pcli, _ := net.Pipe()
	errch := make(chan error, 1)
	upstream := stratum.NewClient(pcli, errch)
	ctx := upstream.Context()
	ctx.ExtraNonce1 = "08000002"
	ctx.ExtraNonce2Size = 4

	list := birpc.List{
		"bf",
		"4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
		birpc.List{},
		"00000002",
		"1c2ac4af",
		"504e86b9",
		false,
	}
	newJob, _ := stratum.NewJob(list)
	ctx.JobCh <- newJob

	p, _ := stratum.NewPoolWithConn(server, order, upstream, errch)
	server.ActivePool(order, p, errch)
	// _ = p.Context()
}

func closeServer() {
	server.Shutdown()
	cli.Close()
	srv.Close()
}

func waitJobChannel(t *testing.T, ctx *stratum.ClientContext) *stratum.Job {
	timeout := time.Duration(100) * time.Millisecond
	var j *stratum.Job

	select {
	case j = <-ctx.JobCh:
		break
	case <-time.After(timeout):
		t.Fatalf("mining.notify timtout.")
	}

	return j
}

func TestSubscribe(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	if client.Active != true {
		t.Fatalf("Client not active.")
	}

	closeServer()
}

func TestSubscribeTimeout(t *testing.T) {
	initServer()
	addOrder()

	time.Sleep(150 * time.Millisecond)

	_, err := io.WriteString(cli, MINING_SUBSCRIBE)
	if err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatalf("client should timeout without subscribe: %v", err)
	}

	closeServer()
}

func TestSubscribeNoPool(t *testing.T) {
	initServer()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err == nil || err.Error() != "No pool available" {
		t.Fatalf("Should not have pools available: %v", err)
	}

	closeServer()
}

func TestSetDifficulty(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	if client.Active != true {
		t.Fatalf("Client not active.")
	}

	ctx := client.Context()
	curJob := waitJobChannel(t, ctx)
	if ctx.Difficulty != stratum.DefaultDifficulty {
		t.Fatalf("mining.set_difficulty not received.")
	}
	if curJob.JobId != "bf" {
		t.Fatalf("mining.notify not received.")
	}

	closeServer()
}

func TestAuthorize(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	ctx := client.Context()
	err = client.Authorize("1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1", "x")
	if ctx.Authorized != true {
		t.Fatalf("Failed on authorize")
	}

	closeServer()
}

func TestBadAuthorize(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	ctx := client.Context()
	err = client.Authorize("12HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1", "x")
	if ctx.Authorized != false {
		t.Fatalf("mining authorize should fail")
	}

	_, err = io.WriteString(cli, "FAKE")
	if err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatalf("client should closed on authorization fail: %v", err)
	}

	closeServer()
}

func TestSubmit(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	ctx := client.Context()
	err = client.Authorize("1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1", "x")
	if !ctx.Authorized {
		t.Fatalf("mining authorize failed")
	}

	curJob := waitJobChannel(t, ctx)
	if ctx.Difficulty != stratum.DefaultDifficulty {
		t.Fatalf("mining.set_difficulty not received.")
	}

	// real log of miner-pool communication which solved testnet3 block 000000002076870fe65a2b6eeed84fa892c0db924f1482243a6247d931dcab32
	err = client.Submit(ctx.Username, curJob.JobId,
		"0001", "504e86ed", "b2957c02")
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = client.Submit(ctx.Username, curJob.JobId,
		"0001", "504e86ed", "b2957c02")
	err2 := err.(*birpc.Error)
	if err2 == nil || err2.Code != stratum.ErrorDuplicateShare {
		t.Fatalf("duplicated share got accepted.")
	}

	closeServer()
}

func TestNewJob(t *testing.T) {
	initServer()
	addOrder()

	errch := make(chan error)
	client := stratum.NewClient(cli, errch)

	err := client.Subscribe()
	if err != nil {
		t.Fatalf("Failed on subscribe: %v", err)
	}

	ctx := client.Context()
	err = client.Authorize("1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1", "x")
	if !ctx.Authorized {
		t.Fatalf("mining authorize failed")
	}

	curJob := waitJobChannel(t, ctx)
	if curJob.JobId != "bf" {
		t.Fatalf("mining.notify not received.")
	}

	prevJobId := curJob.JobId
	// push newjob
	orderId := 1
	pool, ok := stratum.FindPool(orderId)
	if !ok {
		t.Fatalf("pool not found.")
	}
	upstramCtx := pool.Context()

	list := birpc.List{
		"foo",
		"4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
		birpc.List{"ea9da84d55ebf07f47def6b9b35ab30fc18b6e980fc618f262724388f2e9c591"},
		"00000002",
		"1c2ac4af",
		"504e86b9",
		false,
	}
	newJob, _ := stratum.NewJob(list)
	upstramCtx.JobCh <- newJob

	curJob = waitJobChannel(t, ctx)
	if curJob.JobId != "foo" {
		t.Fatalf("mining.notify not received.")
	}

	if !newJob.MerkleBranch[0].IsEqual(curJob.MerkleBranch[0]) {
		t.Fatalf("mining.notify: job merkle branch not equal.")
	}

	// shahash.String() are big-endian
	if curJob.MerkleBranch[0].String() == "ea9da84d55ebf07f47def6b9b35ab30fc18b6e980fc618f262724388f2e9c591" {
		t.Fatalf("job merkle branch equals big-endian which should be little-endian.")
	}

	// submit previous job
	err = client.Submit(ctx.Username, prevJobId,
		"0001", "504e86ed", "b2957c02")
	if err != nil {
		t.Fatalf(err.Error())
	}

	closeServer()
}

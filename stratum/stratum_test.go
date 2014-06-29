package stratum_test

import (
	"github.com/yinhm/ninepool/stratum"
	"testing"
)

func TestExtraNonceCounter(t *testing.T) {
	counter := stratum.NewExtraNonceCounter()
	if counter.Size != 4 {
		t.Errorf("incorrect counter size %d != 4", counter.Size)
	}

	if counter.Next() != "08000001" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Next() != "08000002" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Nonce2Size() != 4 {
		t.Errorf("incorrect Nonce2Size")
	}
}

func TestProxyExtraNonceCounter(t *testing.T) {
	counter := stratum.NewProxyExtraNonceCounter("08000001", 2, 2)

	if next := counter.Next(); next != "080000010001" {
		t.Errorf("incorrect next nonce1: %v", next)
	}

	if counter.Next() != "080000010002" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Nonce2Size() != 2 {
		t.Errorf("incorrect Nonce2Size")
	}
}

func TestDecodeNtime(t *testing.T) {
	ntime, err := stratum.DecodeNtime("504e86ed")
	if err != nil || ntime != int64(1347323629) {
		t.Errorf("failed on parse ntime")
	}
}

func TestCoinbaseHash(t *testing.T) {
	coinbase := stratum.CoinbaseHash(
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
		"08000001",
		"0001",
		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
	)
	if coinbase != "94f317184323c9965abd532450519e6db6859b53b0551c6b8702c1f300ec9b51" {
		t.Errorf("failed to build coinbase %s", coinbase)
	}
}

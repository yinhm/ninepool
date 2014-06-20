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

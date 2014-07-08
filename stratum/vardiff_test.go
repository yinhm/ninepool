package stratum_test

import (
	"github.com/yinhm/ninepool/stratum"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	r := stratum.NewRingFloat64(5)

	if r.Len() != 5 {
		t.Errorf("size %d != 5", r.Len())
	}

	if r.Avg() != 0.0 {
		t.Errorf("avg %.2f != 0.0", r.Avg())
	}

	r.Append(100.0)
	if r.Avg() != 100.0 {
		t.Errorf("avg %.2f", r.Avg())
	}

	r.Append(200.0)
	r.Append(300.0)
	r.Append(400.0)
	r.Append(500.0)
	if r.Avg() != 300.0 {
		t.Errorf("avg %.2f", r.Avg())
	}

	r.Append(600.0)
	if r.Avg() != 400.0 {
		t.Errorf("avg %.2f", r.Avg())
	}

	r.Clear()
	if r.Avg() != 0.0 {
		t.Errorf("clear failed: avg %.2f != 0.0", r.Avg())
	}

	if r.Len() != 5 {
		t.Errorf("size %d != 5", r.Len())
	}
}

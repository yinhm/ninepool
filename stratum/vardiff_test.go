package stratum_test

import (
	"github.com/yinhm/ninepool/stratum"
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	r, _ := stratum.NewRingFloat64(5)

	if r.Len() != 5 {
		t.Errorf("len %d != 5", r.Len())
	}

	if r.Avg() != 0.0 {
		t.Errorf("avg %.2f != 0.0", r.Avg())
	}

	r.Append(100.0)
	if r.Avg() != 100.0 {
		t.Errorf("avg %.2f", r.Avg())
	}
	if r.Size() != 1 {
		t.Errorf("size %d != 1", r.Size())
	}

	r.Append(200.0)
	r.Append(300.0)
	r.Append(400.0)
	r.Append(500.0)
	if r.Size() != 5 {
		t.Errorf("size %d != 5", r.Size())
	}
	if r.Avg() != 300.0 {
		t.Errorf("avg %.2f", r.Avg())
	}

	r.Append(600.0)
	if r.Size() != 5 {
		t.Errorf("size %d != 5", r.Size())
	}
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

func TestVarDiffConfig(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 60, 50)
	if config.BufferSize() != 24 {
		t.Errorf("buffer size %d != 24", config.BufferSize())
	}

	if config.TargetMin != 5 {
		t.Errorf("target min seconds %d != 5", config.TargetMin)
	}

	if config.TargetMax != 15 {
		t.Errorf("target max seconds %d != 15", config.TargetMax)
	}

	if config.TargetDuration != 10 {
		t.Errorf("fail to init target duration")
	}

	if config.RetargetDuration != 60 {
		t.Errorf("fail to init retarget duration")
	}
}

func TestNewVarDiff(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 60, 50)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("faile to init: %v", err)
	}

	vardiff.Submit(1.0)
	if vardiff.BufferSize() != 1 {
		t.Errorf("fail to submit")
	}

	newdiff := vardiff.Submit(1.0)
	if vardiff.BufferSize() != 2 {
		t.Errorf("fail to submit")
	}

	if newdiff != 1.0 {
		t.Errorf("newdiff != 1.0")
	}
}

func TestVarDiffAdj(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 50, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// cheat it to retarget faster with a larger buffer
	config.RetargetDuration = 1

	items := make([]int, 21, 21) // weied, need +1?
	for _, _ = range items {
		vardiff.Submit(8.0)
	}

	if vardiff.BufferSize() < 20 {
		t.Errorf("buffer not full, size: %d", vardiff.BufferSize())
	}

	time.Sleep(1 * time.Second)

	newdiff := vardiff.Submit(8.0)
	if newdiff == 8.0 {
		t.Errorf("retarget failed")
	}
}

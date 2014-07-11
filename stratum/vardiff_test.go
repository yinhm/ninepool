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

	vardiff.Submit(time.Now(), 1.0)
	if vardiff.BufferSize() != 1 {
		t.Errorf("fail to submit")
	}

	newdiff := vardiff.Submit(time.Now(), 1.0)
	if vardiff.BufferSize() != 2 {
		t.Errorf("fail to submit")
	}

	if newdiff != 1.0 {
		t.Errorf("newdiff != 1.0")
	}
}

func TestVarDiffNotChange(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 50, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// 10 sec per share, equal to target
	items := make([]int, 21, 21) // weied, need +1?
	ts := time.Now().Add(-time.Duration(200) * time.Second)
	for _, _ = range items {
		ts = ts.Add(time.Duration(10) * time.Second)
		vardiff.Submit(ts, 8.0)
	}

	if vardiff.BufferSize() < 20 {
		t.Errorf("buffer not full, size: %d", vardiff.BufferSize())
	}

	newdiff := vardiff.Submit(time.Now(), 8.0)
	if newdiff != 8.0 {
		t.Errorf("retarget failed, %.2f", newdiff)
	}
}

func TestVarDiffDouble(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 100, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// 5 sec per share, double
	diff := 8.0
	items := make([]int, 20, 20) // weied, need +1?
	ts := time.Now()
	for _, _ = range items {
		ts = ts.Add(time.Duration(5) * time.Second)
		diff = vardiff.Submit(ts, diff)
	}

	if diff != 16.0 {
		t.Errorf("retarget failed, %.2f", diff)
	}

	for _, _ = range items {
		ts = ts.Add(time.Duration(5) * time.Second)
		diff = vardiff.Submit(ts, diff)
	}

	if diff != 32.0 {
		t.Errorf("retarget failed, %.2f", diff)
	}
}

func TestVarDiffHalf(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 100, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// 20 sec / share, down half
	diff := 8.0
	olddiff := diff
	items := make([]int, 10, 10)
	ts := time.Now()
	for _, _ = range items {
		ts = ts.Add(time.Duration(20) * time.Second)
		diff = vardiff.Submit(ts, diff)
		if diff != olddiff && diff == 4.0 {
			return
		}
	}

	t.Errorf("retarget failed, %.2f", diff)
}

func TestVarDiffMax(t *testing.T) {
	config := stratum.NewVarDiffConfig(1, 512.0, 10, 100, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// 20 sec / share, down half
	diff := 8.0
	olddiff := diff
	items := make([]int, 80, 80)
	ts := time.Now()
	ts = ts.Add(time.Duration(45) * time.Second)
	for _, _ = range items {
		ts = ts.Add(time.Duration(100) * time.Millisecond)
		diff = vardiff.Submit(ts, diff)
		if diff != olddiff && diff == 512.0 {
			return
		}
	}

	t.Errorf("retarget failed, %.2f", diff)
}

func TestVarDiffLessThanOne(t *testing.T) {
	config := stratum.NewVarDiffConfig(0.001, 0.1, 10, 100, 10)

	vardiff, err := stratum.NewVarDiff(config)
	if err != nil {
		t.Fatalf("fail to init: %v", err)
	}

	// 5 sec / share, double
	diff := 0.01
	olddiff := diff
	items := make([]int, 80, 80)
	ts := time.Now()
	for _, _ = range items {
		ts = ts.Add(time.Duration(5) * time.Second)
		diff = vardiff.Submit(ts, diff)
		if diff != olddiff && diff == 0.02 {
			return
		}
	}

	t.Errorf("retarget failed, %.2f", diff)
}

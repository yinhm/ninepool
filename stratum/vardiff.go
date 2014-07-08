package stratum

import (
	"container/ring"
	"time"
)

type RingFloat64 struct {
	max  int
	ring *ring.Ring
}

func NewRingFloat64(n int) *RingFloat64 {
	r := ring.New(n)
	return &RingFloat64{max: n, ring: r}
}

func (r *RingFloat64) Append(x float64) {
	r.ring.Value = x
	r.ring = r.ring.Next()
}

func (r *RingFloat64) Avg() float64 {
	sum := 0.0
	size := 0
	r.ring.Do(func(p interface{}) {
		if p != nil {
			size++
			sum += p.(float64)
		}
	})
	if size == 0 {
		return 0.0
	}
	return sum / float64(size)
}

func (r *RingFloat64) Len() int {
	return r.ring.Len()
}

func (r *RingFloat64) Size() int {
	size := 0
	r.ring.Do(func(p interface{}) {
		if p != nil {
			size++
		}
	})
	return size
}

func (r *RingFloat64) Clear() {
	r.ring = ring.New(r.max)
}

type VarDiffConfig struct {
	min              float64
	max              float64
	tMin             int // target min
	tMax             int // target max
	targetDuration   int64
	retargetDuration int64
	x2mode           bool
}

type vardiff struct {
	config       *VarDiffConfig
	timeBuffer   *RingFloat64
	retargetTime time.Time
	updated      time.Time
}

// eg: NewVarDiff(16.0, 512.0, 15, 90, 30)
func NewVarDiff(min, max float64, target, retarget, variance int) *vardiff {
	varLimit := int(float64(target) * (float64(variance) / 100.0))
	tMin := target - varLimit
	tMax := target + varLimit

	config := &VarDiffConfig{
		min:              0.1,
		max:              1024.0,
		tMin:             tMin,
		tMax:             tMax,
		targetDuration:   int64(target),
		retargetDuration: int64(retarget),
		x2mode:           true,
	}

	ts := time.Now()
	retargetTime := ts.Add(-time.Duration(config.retargetDuration/2) * time.Second)
	bufferSize := retarget / target * 4

	return &vardiff{
		config:       config,
		timeBuffer:   NewRingFloat64(bufferSize),
		retargetTime: retargetTime,
		updated:      ts,
	}
}

// Submit shares, calcuate new difficulty.
func (v *vardiff) Submit(worker *Worker) {
	ts := time.Now()

	// log last share work time
	v.timeBuffer.Append(ts.Sub(v.updated).Seconds())
	v.updated = ts

	sinceRetarget := int64(ts.Sub(v.retargetTime).Seconds())
	// no need to retarget
	if sinceRetarget < v.config.retargetDuration && v.timeBuffer.Size() > 0 {
		return
	}

	v.retargetTime = ts
	avg := int(v.timeBuffer.Avg())
	ddiff := float64(v.config.targetDuration) / float64(avg)

	workerDiff := worker.context.Difficulty
	if avg > v.config.tMax && workerDiff > v.config.min {
		// lower diff, more shares
		if v.config.x2mode {
			ddiff = 0.5
		}
	} else if avg < v.config.tMin {
		// increase diff, less shares
		if v.config.x2mode {
			ddiff = 2
		}
	}

	newDiff := workerDiff * ddiff
	if newDiff < v.config.min {
		newDiff = v.config.min
	}
	if newDiff > v.config.max {
		newDiff = v.config.max
	}

	v.timeBuffer.Clear()
	worker.newDifficulty(newDiff)
}

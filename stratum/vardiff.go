package stratum

import (
	"container/ring"
	"errors"
	"time"
)

type RingFloat64 struct {
	max  int
	ring *ring.Ring
}

func NewRingFloat64(n int) (*RingFloat64, error) {
	// lower n makes this pointless
	if n < 5 {
		return nil, errors.New("too low ring buffer maxsize.")
	}
	r := ring.New(n)
	return &RingFloat64{max: n, ring: r}, nil
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
	x2mode           bool
	min              float64
	max              float64
	TargetMin        int // target min
	TargetMax        int // target max
	TargetDuration   int
	RetargetDuration int
}

func NewVarDiffConfig(min, max float64, target, retarget, variance int) *VarDiffConfig {
	varLimit := int(float64(target) * (float64(variance) / 100.0))
	if varLimit < 1 {
		varLimit = 1
	}
	tMin := target - varLimit
	tMax := target + varLimit

	return &VarDiffConfig{
		x2mode:           false,
		min:              min,
		max:              max,
		TargetMin:        tMin,
		TargetMax:        tMax,
		TargetDuration:   target,
		RetargetDuration: retarget,
	}
}

func (c *VarDiffConfig) BufferSize() int {
	return 4 * c.RetargetDuration / c.TargetDuration
}

type vardiff struct {
	config       *VarDiffConfig
	timeBuffer   *RingFloat64
	retargetTime time.Time
	updated      time.Time
}

// eg: NewVarDiff(16.0, 512.0, 15, 90, 30)
func NewVarDiff(config *VarDiffConfig) (*vardiff, error) {
	ts := time.Now()
	retargetTime := ts.Add(-time.Duration(config.RetargetDuration/2) * time.Second)
	buf, err := NewRingFloat64(config.BufferSize())
	if err != nil {
		return nil, err
	}
	d := &vardiff{
		config:       config,
		timeBuffer:   buf,
		retargetTime: retargetTime,
		updated:      ts,
	}
	return d, nil
}

// Submit shares, calcuate new difficulty.
func (v *vardiff) Submit(shareTime time.Time, oldDiff float64) float64 {
	// log last share work time
	v.timeBuffer.Append(shareTime.Sub(v.updated).Seconds())
	v.updated = shareTime

	sinceRetarget := int(shareTime.Sub(v.retargetTime).Seconds())
	// no need to retarget
	// log.Printf("sinceRetarget = %d, v.config.RetargetDuration = %d",
	//  sinceRetarget, v.config.RetargetDuration)
	if sinceRetarget < v.config.RetargetDuration && v.BufferSize() > 0 {
		return oldDiff
	}

	v.retargetTime = shareTime
	return v.calculate(oldDiff)
}

func (v *vardiff) calculate(oldDiff float64) float64 {
	avg := int(v.timeBuffer.Avg())
	ddiff := float64(v.config.TargetDuration) / float64(avg)

	if avg > v.config.TargetMax && oldDiff > v.config.min {
		// lower diff, more shares
		if v.config.x2mode {
			ddiff = 0.5
		}
	} else if avg < v.config.TargetMin {
		// increase diff, less shares
		if v.config.x2mode {
			ddiff = 2
		}
	}

	newDiff := oldDiff * ddiff

	if newDiff < v.config.min {
		newDiff = v.config.min
	}
	if newDiff > v.config.max {
		newDiff = v.config.max
	}

	v.timeBuffer.Clear()
	return newDiff
}

func (v *vardiff) BufferSize() int {
	return v.timeBuffer.Size()
}

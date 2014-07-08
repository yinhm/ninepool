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

func (r *RingFloat64) Clear() {
	r.ring = ring.New(r.max)
}


type vardiff struct {
	bSize int
	tMin  int
	tMax  int
	retarget int64
}

// eg: NewVarDiff(16.0, 512.0, 15, 90, 30)
func NewVardiff(min, max float64, target, retarget, variance int) *vardiff {
  varLimit := int(float64(target) * (float64(variance) / 100.0))
	bufferSize := retarget / target * 4
	tMin := target - varLimit
	tMax := target + varLimit

	return &vardiff {
		bSize: bufferSize,
		tMin: tMin,
		tMax: tMax,
		retarget: int64(retarget),
	}
}

// Submit shares, calcuate new difficulty.
func (v *vardiff) Submit(client *Worker) float64 {
  // var lastTs int64
  var lastRtc int64

  ts := time.Now().Unix()

  if lastRtc != 0.0 {
    lastRtc = ts - v.retarget / 2
    // lastTs = ts
    timeBuffer := NewRingFloat64(v.bSize)
		return timeBuffer.Avg()
  }
  return 0.0
}

package profiler

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// MaxTryCntSwap - max try const swap
const MaxTryCntSwap = 3

var m sync.Mutex
var mapProfiler map[string]*profiler

// Profiler interface
type Profiler interface {
	Start() time.Time
	End(time.Time) time.Duration
	Info() string
}

type profiler struct {
	Name string
	// By priority atomics operations  from max >> to min
	CntStart uint64
	CntEnd   uint64
	SumTime  uint64
	MinTime  uint64
	MaxTime  uint64
}

func init() {
	mapProfiler = map[string]*profiler{}
}

// GetProfiler - get profiler instance
func GetProfiler(name string) Profiler {
	emptyProfiler := &profiler{}
	emptyProfiler.MinTime = math.MaxUint64

	if name == "" {
		return emptyProfiler
	}

	emptyProfiler.Name = name

	m.Lock()
	defer m.Unlock()

	if profiler, ok := mapProfiler[name]; ok {
		return profiler
	}
	mapProfiler[name] = emptyProfiler
	return emptyProfiler
}

// Start -
func (p *profiler) Start() time.Time {
	atomic.AddUint64(&p.CntStart, 1)
	return time.Now()
}

// End -
func (p *profiler) End(startTime time.Time) time.Duration {
	atomic.AddUint64(&p.CntEnd, 1)
	delta := time.Now().Sub(startTime)
	deltaUint64 := uint64(delta)

	atomic.AddUint64(&p.SumTime, uint64(delta))

	for i := 0; i < MaxTryCntSwap; i++ {
		if minTime := p.MinTime; deltaUint64 < minTime {
			if atomic.CompareAndSwapUint64(&p.MinTime, minTime, deltaUint64) {
				break
			}
		}
	}

	for i := 0; i < MaxTryCntSwap; i++ {
		if maxTime := p.MaxTime; deltaUint64 > maxTime {
			if atomic.CompareAndSwapUint64(&p.MaxTime, maxTime, deltaUint64) {
				break
			}
		}
	}

	return delta
}

func (p *profiler) String() string {
	return fmt.Sprintf(
		"\nName: %v"+
			"\nCntStart: %v"+
			"\nCntEnd: %v"+
			"\nSumTime: %v"+
			"\nMinTime: %v"+
			"\nMaxTime: %v\n"+
			"=========\nAvgTime: %v\n",
		p.Name, p.CntStart, p.CntEnd,
		time.Duration(p.SumTime), time.Duration(p.MinTime), time.Duration(p.MaxTime),
		time.Duration(p.SumTime/p.CntEnd))
}

func (p *profiler) Info() string {
	return fmt.Sprintf("%+v", p)
}

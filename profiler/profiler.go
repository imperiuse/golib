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

var (
	mProfiler   sync.Mutex
	mapProfiler map[string]*profiler
	mTimer      sync.Mutex
	mapTimer    map[string]*timer
)

// Profiler - Atomic Profiler interface (no mutex) SAVE FOR GOROUTINE!
type (
	Profiler interface {
		Start() time.Time
		End(time.Time) time.Duration
		Info() string
	}

	// Timer - time save profiler (no mutex) NOT SAFE FOR GOROUTINE!
	Timer interface {
		Start()
		End()
		GetDuration() time.Duration
		GetStartTime() time.Time
		GetEndTime() time.Time
		Info() string
	}

	profiler struct {
		name string
		// By priority atomics operations  from max >> to min
		cntStart uint64
		cntEnd   uint64
		sumTime  uint64
		minTime  uint64
		maxTime  uint64
	}

	timer struct {
		name string

		startTime time.Time
		endTime   time.Time
		duration  time.Duration
	}
)

func init() {
	mapProfiler = map[string]*profiler{}
	mapTimer = map[string]*timer{}
}

// GetProfiler - get profiler instance by name or create new
func GetProfiler(name string) Profiler {
	emptyProfiler := profiler{
		name:    name,
		minTime: math.MaxUint64}

	mProfiler.Lock()
	defer mProfiler.Unlock()

	profiler, ok := mapProfiler[name]
	if !ok {
		mapProfiler[name] = &emptyProfiler
		profiler = &emptyProfiler
	}
	return profiler
}

// GetTimer - get timer instance by name or create new
func GetTimer(name string) Timer {
	emptyTimer := timer{name: name}

	mTimer.Lock()
	defer mTimer.Unlock()

	timer, ok := mapTimer[name]
	if !ok {
		mapTimer[name] = &emptyTimer
		timer = &emptyTimer
	}
	return timer
}

// Start - return Time.Now() and increment count use
func (p *profiler) Start() time.Time {
	atomic.AddUint64(&p.cntStart, 1)
	return time.Now()
}

// End - return time.Duration use `time.Since(startTime)`
func (p *profiler) End(startTime time.Time) time.Duration {
	atomic.AddUint64(&p.cntEnd, 1)
	delta := time.Since(startTime)
	deltaUint64 := uint64(delta)

	atomic.AddUint64(&p.sumTime, uint64(delta))

	for i := 0; i < MaxTryCntSwap; i++ {
		minTime := p.minTime
		if deltaUint64 < minTime {
			if atomic.CompareAndSwapUint64(&p.minTime, minTime, deltaUint64) {
				break
			}
		}
	}

	for i := 0; i < MaxTryCntSwap; i++ {
		maxTime := p.maxTime
		if deltaUint64 > maxTime {
			if atomic.CompareAndSwapUint64(&p.maxTime, maxTime, deltaUint64) {
				break
			}
		}
	}

	return delta
}

// String - pretty print
func (p *profiler) String() string {
	return fmt.Sprintf(
		"\nName: %v"+
			"\nCntStart: %v"+
			"\nCntEnd: %v"+
			"\nSumTime: %v"+
			"\nMinTime: %v"+
			"\nMaxTime: %v\n"+
			"=========\nAvgTime: %v\n",
		p.name, p.cntStart, p.cntEnd,
		time.Duration(p.sumTime), time.Duration(p.minTime), time.Duration(p.maxTime),
		time.Duration(p.sumTime/p.cntEnd))
}

// Info - profiler stats
func (p *profiler) Info() string {
	return p.String()
}

// Start - start timer
func (t *timer) Start() {
	mTimer.Lock()
	defer mTimer.Unlock()

	t.startTime = time.Now()
}

// End - stop timer
func (t *timer) End() {
	t.endTime = time.Now()
	t.duration = t.endTime.Sub(t.startTime)
}

// Duration - getter duration
func (t *timer) GetDuration() time.Duration {
	return t.duration
}

// GetStartTime - getter startTime
func (t *timer) GetStartTime() time.Time {
	return t.startTime
}

// GetEndTime - getter endTime
func (t *timer) GetEndTime() time.Time {
	return t.endTime
}

// String - pretty print
func (t *timer) String() string {
	return fmt.Sprintf("\nName:%s\nStartTime:%v\nEndTime:%v\nDuration:%v\n", t.name, t.startTime, t.endTime, t.duration)
}

// Info - timer stats
func (t *timer) Info() string {
	return t.String()
}

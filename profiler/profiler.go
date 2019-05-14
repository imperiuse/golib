package profiler

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/k0kubun/pp"
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
type Profiler interface {
	Start() time.Time
	End(time.Time) time.Duration
	Info() string
}

// Timer - time save profiler (no mutex) NOT SAFE FOR GOROUTINE!
type Timer interface {
	Start()
	End()
	Duration() time.Duration
	Info() string
}

// profiler -
type profiler struct {
	name string
	// By priority atomics operations  from max >> to min
	cntStart uint64
	cntEnd   uint64
	sumTime  uint64
	minTime  uint64
	maxTime  uint64
}

// timer -
type timer struct {
	startTime time.Time
	endTime   time.Time
	duration  time.Duration
}

func init() {
	mapProfiler = map[string]*profiler{}
	mapTimer = map[string]*timer{}
}

// GetProfiler - get profiler instance
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

// GetTimer - get timer instance
func GetTimer(name string) Timer {
	emptyTimer := timer{}

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

// End -
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

// String -
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

// Info -
func (p *profiler) Info() string {
	return p.String()
}

func (t *timer) Start() {
	mTimer.Lock()
	defer mTimer.Unlock()

	t.startTime = time.Now()
}

func (t *timer) End() {
	t.endTime = time.Now()
	t.duration = t.endTime.Sub(t.startTime)
}

func (t *timer) Duration() time.Duration {
	return t.duration
}

func (t *timer) Info() string {
	return pp.Sprintln(t)
}

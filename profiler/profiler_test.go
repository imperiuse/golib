package profiler

import (
	"testing"
	"time"
)

func TestProfiler_1(t *testing.T) {
	p := GetProfiler("")
	tt := p.Start()
	time.Sleep(time.Second)
	p.End(tt)

	t.Log(p.Info())
}

func TestProfiler_2(t *testing.T) {
	p1 := GetProfiler("1")
	tt := p1.Start()
	time.Sleep(time.Second)
	p1.End(tt)
	t.Log(p1.Info())

	p2 := GetProfiler("1")
	t.Log(p2.Info())

	if p1 != p2 {
		t.Error("Pointers obj profiler different!  p1 != p2")
	}
}

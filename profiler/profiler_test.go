package profiler

import (
	"testing"
	"time"
)

func TestProfiler_1(T *testing.T) {
	p := GetProfiler("")
	t := p.Start()
	time.Sleep(time.Second)
	p.End(t)
	T.Log(p.Info())
}

func TestProfiler_2(T *testing.T) {
	p1 := GetProfiler("1")
	t := p1.Start()
	time.Sleep(time.Second)
	p1.End(t)
	T.Log(p1.Info())

	p2 := GetProfiler("1")
	T.Log(p2.Info())

	if p2 != p2 {
		T.Error("Pointers obj profiler different!  p1 != p2")
	}
}

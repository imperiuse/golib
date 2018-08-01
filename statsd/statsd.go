package statsd

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Statsd interface {
	Flush() bool

	Sum(string, int) bool
	Inc(string) bool
	Dec(string) bool

	Gauge(string, int) bool
	Timing(string, int) bool

	Close()
}

type commandAction int

const (
	limit         = 512
	sleepMilliSec = 100
)

const (
	modify commandAction = 1 + iota
	flush
)

type statsData struct {
	action commandAction
	key    string
	value  int
	add    string
}

type statsd struct {
	chn      chan statsData
	base     string
	conn     net.Conn
	isOpened bool
}

func statsd_flush(conn net.Conn, arr []byte) {
	if len(arr) > 0 {
		conn.Write(arr)
	}
}

func statsd_prepare(base string, data *statsData) string {
	return fmt.Sprintf("%s%s:%d|%s\n", base, data.key, data.value, data.add)
}

func (st statsd) run() {
	store := ""
	for {
		select {
		case command := <-st.chn:
			switch command.action {
			case flush:
				if len(store) > 0 {
					statsd_flush(st.conn, []byte(store))
				}

			case modify:
				cmd := statsd_prepare(st.base, &command)
				if len(store) > 0 && len(store)+len(cmd) > limit {
					statsd_flush(st.conn, []byte(store))
					store = ""
				}
				store = store + cmd
			}

		case <-time.After(time.Duration(sleepMilliSec) * time.Millisecond):
			if !st.isOpened {
				return
			}
		}
	}
}

func (st statsd) Sum(key string, value int) bool {
	if st.isOpened {
		st.chn <- statsData{modify, key, value, "c"}
		return true
	} else {
		return false
	}
}

func (st statsd) Inc(key string) bool {
	return Statsd(st).Sum(key, 1)
}

func (st statsd) Dec(key string) bool {
	return Statsd(st).Sum(key, -1)
}

func (st statsd) Gauge(key string, value int) bool {
	if st.isOpened {
		st.chn <- statsData{modify, key, value, "g"}
		return true
	} else {
		return false
	}
}

func (st statsd) Timing(key string, value int) bool {
	if st.isOpened {
		st.chn <- statsData{modify, key, value, "ms"}
		return true
	} else {
		return false
	}
}

func (st statsd) Flush() bool {
	if st.isOpened {
		st.chn <- statsData{flush, "", 0, ""}
		return true
	} else {
		return false
	}
}

func (st statsd) Close() {
	if st.isOpened {
		st.conn.Close()
		st.isOpened = false
	}
}

func New(bs string, addr string) Statsd {
	if !strings.HasSuffix(bs, ".") {
		bs = bs + "."
	}
	st := statsd{chn: make(chan statsData), base: bs, conn: nil, isOpened: false}
	connection, err := net.Dial("udp", addr)
	if err == nil {
		st.conn = connection
		st.isOpened = true
		go st.run()
	}
	return st
}

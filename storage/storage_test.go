package storage

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestNew(t *testing.T) {
	storage := New(time.Second * 10)
	if storage.m == nil || storage.TTL == 0 {
		t.Errorf("return non initialize storage")
	}

}

func TestSetAndGet(t *testing.T) {

	testCases := [][]string{
		{"test.domain.ru", strings.Repeat("very big string", 1000)},
		{"test.domain.ru", strings.Repeat("small string", 100)},
		{"test1.domain.ru", strings.Repeat("very big string", 10000)},
		{"test2.domain.ru", strings.Repeat("very big string", 20000)},
	}

	storage := New(time.Second * 10)

	for _, v := range testCases {
		storage.Set(v[0], v[1])
	}

	for i, v := range testCases {
		if s, found := storage.Get(v[0]); !found {
			t.Errorf("not found whois text for domain %s test case #%d", v[0], i)

		} else {
			if i != 0 {
				if s != v[1] {
					t.Errorf("get text non equals set text for domain %s test case #%d", v[0], i)
				}
			} else { // проверка что мы перезаписали старый Set
				if s == v[0] {
					t.Errorf("get text non equals set text for domain %s test case #%d", v[0], i)
				}
			}
		}
	}
}

func TestTTLnegative(t *testing.T) {
	const TTL = time.Second * 1
	storage := New(TTL)

	testCases := [][]string{
		{"test.domain.ru", strings.Repeat("small string", 100)},
		{"test1.domain.ru", strings.Repeat("big string", 1000)},
		{"test2.domain.ru", strings.Repeat("very big string", 10000)},
	}

	for _, v := range testCases {
		storage.Set(v[0], v[1])
	}

	time.Sleep(TTL)

	for i, v := range testCases {
		if _, found := storage.Get(v[0]); found {
			t.Errorf("not remove (TTL) whois text for domain %s test case #%d", v[0], i)

		}
	}

}

func TestGoRoutinSafe(t *testing.T) {
	const (
		NumberGoroutine = 50
		domain          = "test.test"
		TTL             = time.Millisecond * 250
	)

	storage := New(TTL)
	storage.Set(domain, "init whois info")

	stop := make(chan interface{}, 1)

	for i := 0; i < NumberGoroutine; i++ {
		go func(i int) {
			sr := rand.NewSource(time.Now().UnixNano())
			r := rand.New(sr)
			whoisText := fmt.Sprintf("%d: whois info", i)
			for {
				select {
				case <-stop:
					return
				default:
					rw := r.Intn(2)
					if rw == 0 {
						_, found := storage.Get(domain)
						if !found {
							t.Errorf("!found")
						}
					} else {
						storage.Set(domain, whoisText)
					}
					time.Sleep(time.Microsecond * 10 * time.Duration(r.Intn(5)))
				}
			}
		}(i)
	}

	time.Sleep(time.Second * 5)

	close(stop)

}

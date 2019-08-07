package storage

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestStorage_New(t *testing.T) {
	storage := New(time.Second*10, time.Hour*24)
	if storage == nil || storage.m == nil || storage.TTL == 0 {
		t.Errorf("return non initialize storage")
	}

}

func TestStorage_New_Stupid(t *testing.T) {
	storage := New(0, 0)
	if storage == nil || storage.m == nil || storage.TTL == 0 {
		t.Errorf("return non initialize storage")
	}

}

func TestStorage_SetAndGet(t *testing.T) {

	testCases := []struct {
		key   string
		value interface{}
	}{
		{"key1", strings.Repeat("very big string", 100)},
		{"key1", strings.Repeat("small string", 100)},
		{"key3", strings.Repeat("very big string", 100)},
		{"key4", strings.Repeat("very big string", 100)},
	}

	storage := New(time.Second*10, time.Hour*24)

	for _, v := range testCases {
		storage.Set(v.key, v.value)
	}

	for i, v := range testCases {
		if s, found := storage.Get(v.key); !found {
			t.Errorf("not found data for key %s test case #%d", v.key, i)

		} else {
			if i != 0 {
				if s != v.value {
					t.Errorf("get text non equals set data for key %s test case #%d", v.key, i)
				}
			} else { // проверка что мы перезаписали старый Set
				if s == v.value {
					t.Error(s, v.value)
					t.Errorf("get text non equals set data for key %s test case #%d", v.key, i)
				}
			}
		}
	}
}

func TestStorage_Delete(t *testing.T) {

	testCases := []struct {
		key   string
		value interface{}
	}{
		{"key1", 1234},
		{"key2", "1234"},
		{"key3", struct{}{}},
		{"key4", nil},
	}

	storage := New(time.Second*10, time.Hour*24)

	for _, v := range testCases {
		storage.Set(v.key, v.value)
		storage.Delete(v.key)
	}

	for i, v := range testCases {
		if _, found := storage.Get(v.key); found {
			t.Errorf("found deleted key: %s test case #%d", v.key, i)

		}
	}
}

func TestStorage_RemoveAll(t *testing.T) {

	testCases := [][]string{
		{"key1", strings.Repeat("very big string", 1000)},
		{"key2", strings.Repeat("small string", 100)},
		{"key3", strings.Repeat("very big string", 10000)},
		{"key4", strings.Repeat("very big string", 20000)},
	}

	storage := New(time.Second*10, time.Hour*24)

	for _, v := range testCases {
		storage.Set(v[0], v[1])
	}

	storage.RemoveAll()

	for i, v := range testCases {
		if _, found := storage.Get(v[0]); found {
			t.Fatalf("found data for key %s test case #%d", v[0], i)
		}
	}

}

func TestStorage_CacheTTL_negative(t *testing.T) {
	const TTL = time.Second * 1
	storage := New(TTL, time.Hour*24)

	testCases := [][]string{
		{"key1", strings.Repeat("small string", 100)},
		{"key2", strings.Repeat("big string", 1000)},
		{"key3", strings.Repeat("very big string", 10000)},
	}

	for _, v := range testCases {
		storage.Set(v[0], v[1])
	}

	time.Sleep(TTL)

	for i, v := range testCases {
		if _, found := storage.Get(v[0]); found {
			t.Errorf("not remove (TTL) data for key %s test case #%d", v[0], i)

		}
	}

}

func TestStorage_ResetCache(t *testing.T) {
	const TTL = time.Second * 100
	storage := New(TTL, time.Second*2)

	testCases := [][]string{
		{"key1", strings.Repeat("small string", 100)},
		{"key2", strings.Repeat("big string", 1000)},
		{"key3", strings.Repeat("very big string", 10000)},
	}

	for _, v := range testCases {
		storage.Set(v[0], v[1])
	}

	time.Sleep(time.Second * 2)

	storage.RemoveAll()

	for i, v := range testCases {
		if _, found := storage.Get(v[0]); found {
			t.Errorf("not remove (ResetCache) data for key %s test case #%d", v[0], i)
		}
	}
}

func TestStorage_GoRoutinSafe(t *testing.T) {
	const (
		NumberGoroutine = 50
		key             = "key"
		TTL             = time.Millisecond * 250
	)

	storage := New(TTL, time.Hour*24)
	storage.Set(key, "init info")

	stop := make(chan interface{}, 1)

	for i := 0; i < NumberGoroutine; i++ {
		go func(i int) {
			sr := rand.NewSource(time.Now().UnixNano())
			r := rand.New(sr)
			data := fmt.Sprintf("%d: info", i)
			for {
				select {
				case <-stop:
					return
				default:
					rw := r.Intn(3)
					if rw == 0 {
						_, _ = storage.Get(key)
					} else if rw == 1 {
						storage.Set(key, data)
					} else {
						storage.Delete(key)
					}
					time.Sleep(time.Microsecond * 10 * time.Duration(r.Intn(5)))
				}
			}
		}(i)
	}

	time.Sleep(time.Second * 5)

	close(stop)

}

package cache

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStorage_New(t *testing.T) {
	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestStorage_New2(t *testing.T) {
	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestStorage_New_Bad(t *testing.T) {
	s, err := New[string, any](Config{
		DefaultTTL: "qwrqwrfqw",
	})

	assert.NotNil(t, err)
	assert.Nil(t, s)
}

func TestStorage_SetAndGet(t *testing.T) {
	testCases := []struct {
		key   string
		value any
	}{
		{"key1", strings.Repeat("very big string", 100)},
		{"key1", strings.Repeat("small string", 100)},
		{"key3", strings.Repeat("very big string", 100)},
		{"key4", strings.Repeat("very big string", 100)},
	}

	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	for _, v := range testCases {
		s.Set(v.key, v.value)
	}

	for i, v := range testCases {
		value, found := s.Get(v.key)

		assert.True(t, found)

		if i != 0 {
			assert.Equal(t, value, v.value)
		} else { // проверка что мы перезаписали старый Set
			assert.NotEqual(t, value, v.value)
		}
	}
}

func TestStorage_TryGetOrInvokeLambda(t *testing.T) {
	testCases := []struct {
		key   string
		value any
	}{
		{"key1", 1234},
		{"key2", "1234"},
		{"key3", struct{}{}},
		{"key4", nil},
	}

	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	for _, v := range testCases {
		s.Set(v.key, v.value)

		val, err := s.TryGetOrInvokeLambda(v.key, nil)
		assert.Nil(t, err)
		assert.Equal(t, v.value, val)
	}

	var mockErr = errors.New("test_mock_err")
	val, err := s.TryGetOrInvokeLambda("not_exist", func(key string) (any, TTL, error) {
		return nil, TTL{}, mockErr
	})
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, mockErr))
	assert.Nil(t, val)

	var (
		d   = "data"
		ttl = TTL{TTL: time.Second, ExpireAt: time.Time{}}
	)
	val, err = s.TryGetOrInvokeLambda("not_exist_yet", func(key string) (any, TTL, error) {
		return d, ttl, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, d, val)

	val, found := s.Get("not_exist_yet")
	assert.True(t, found)
	assert.Equal(t, d, val)

	time.Sleep(time.Second)

	val, found = s.Get("not_exist_yet")
	assert.False(t, found)
	assert.Nil(t, val)

}

func TestStorage_Delete(t *testing.T) {
	testCases := []struct {
		key   string
		value any
	}{
		{"key1", 1234},
		{"key2", "1234"},
		{"key3", struct{}{}},
		{"key4", nil},
	}

	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	for _, v := range testCases {
		s.Set(v.key, v.value)
		s.Delete(v.key)
	}

	for _, v := range testCases {
		_, found := s.Get(v.key)
		assert.False(t, found)
	}
}

func TestStorage_RemoveAll(t *testing.T) {
	testCases := []struct {
		key   string
		value any
	}{
		{"key1", strings.Repeat("very big string", 1000)},
		{"key2", strings.Repeat("small string", 100)},
		{"key3", strings.Repeat("very big string", 10000)},
		{"key4", strings.Repeat("very big string", 20000)},
	}

	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	for _, v := range testCases {
		s.Set(v.key, v.value)
	}

	s.CleanAll()

	for _, v := range testCases {
		_, found := s.Get(v.key)
		assert.False(t, found)
	}

}

func TestStorage_CacheTTL_negative(t *testing.T) {
	const ttlSecond = time.Second * 1

	s, err := New[string, any](Config{
		DefaultTTL: "60s",
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	testCases := []struct {
		key   string
		value any
	}{
		{"key1", strings.Repeat("very big string", 100)},
		{"key2", strings.Repeat("small string", 10)},
		{"key3", strings.Repeat("very big string", 1000)},
	}

	for _, v := range testCases {
		s.SetWithTTL(v.key, v.value, TTL{TTL: ttlSecond})

		s.SetWithTTL(v.key+"_", v.value, TTL{ExpireAt: time.Now().Add(ttlSecond)})

		s.SetWithTTL(v.key+"__", v.value, TTL{TTL: ttlSecond, ExpireAt: time.Now().Add(ttlSecond)})

		s.Set(v.key+"+", v.value)
		s.SetTTL(v.key+"+", TTL{TTL: ttlSecond})

		s.Set(v.key+"++", v.value)
		s.SetTTL(v.key+"++", TTL{ExpireAt: time.Now().Add(ttlSecond)})

		s.Set(v.key+"+++", v.value)
		s.SetTTL(v.key+"+++", TTL{TTL: ttlSecond, ExpireAt: time.Now().Add(ttlSecond)})
	}
	s.SetTTL("no_exist", TTL{TTL: ttlSecond})
	s.SetTTL("no_exist", TTL{ExpireAt: time.Now().Add(ttlSecond)})
	s.SetTTL("no_exist", TTL{TTL: ttlSecond, ExpireAt: time.Now().Add(ttlSecond)})

	time.Sleep(ttlSecond + time.Millisecond*100)

	for _, v := range testCases {
		_, found := s.Get(v.key)
		assert.False(t, found)

		_, found = s.Get(v.key + "_")
		assert.False(t, found)

		_, found = s.Get(v.key + "__")
		assert.False(t, found)

		_, found = s.Get(v.key + "+")
		assert.False(t, found)

		_, found = s.Get(v.key + "++")
		assert.False(t, found)

		_, found = s.Get(v.key + "+++")
		assert.False(t, found)
	}

}

func TestStorage_GoroutineSafe(t *testing.T) {
	const (
		NumberGoroutine = 50
		key             = "key"
		TTL             = "250ms"
	)

	s, err := New[string, any](Config{
		DefaultTTL: TTL,
	})

	assert.Nil(t, err)
	assert.NotNil(t, s)

	s.Set(key, "init info")

	stop := make(chan any, 1)

	for i := 0; i < NumberGoroutine; i++ {
		go func(i int) {
			sr := rand.NewSource(time.Now().UnixNano())
			r := rand.New(sr)
			d := fmt.Sprintf("%d: info", i)
			for {
				select {
				case <-stop:
					return
				default:
					rw := r.Intn(4)
					if rw == 0 {
						_, _ = s.Get(key)
					} else if rw == 1 {
						s.Set(key, d)
					} else {
						s.Delete(key)
					}
					time.Sleep(time.Microsecond * 10 * time.Duration(r.Intn(5)))
				}
			}
		}(i)
	}

	time.Sleep(time.Second * 5)

	close(stop)

}

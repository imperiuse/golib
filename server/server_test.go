package server

import (
	"net"
	"strconv"
	"testing"
)

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestNew(t *testing.T) {
	server, err := New("tcp", "localhost", "50000", 1)
	if err != nil || server == nil {
		t.Errorf("return non initialize server or erorr: %v", err)
	}
}

func TestNew_Negative(t *testing.T) {
	testCases := [][]string{
		{"", "localhost", "50000", "1"},
		{"tcp", "", "50000", "1"},
		{"tcp", "localhost", "", "1"},
		{"tcp", "localhost", "50000", "0"},
		{"tcp", "", "50000", "0"},
		{"tcp", "localhost", "", "0"},
		{"edqwe", "localhost", "5555", "10"},
	}

	for _, v := range testCases {
		cnt, err := strconv.ParseInt(v[3], 10, 32)
		if err != nil {
			t.Errorf("can't convert to int: %v", v[3])
		}
		_, err = New(ConnType(v[0]), v[1], v[2], int(cnt))
		if err == nil {
			t.Errorf("no err for incorect params!")
		}
	}
}

func TestServer_Addr(t *testing.T) {
	server, _ := New("tcp", "localhost", "50000", 1)

	if server.Addr() != "localhost:50000" {
		t.Error("server.Addr() return wrong answer")
	}
}

func TestServer_ListenAndServe(t *testing.T) {
	server, _ := New("tcp", "localhost", "50000", 1)

	nothingFunc := func(conn net.Conn) error {
		defer func() {
			_ = conn.Close()
		}()

		return nil
	}
	chErr := make(chan error, 1)

	err := server.ListenAndServe(nothingFunc, chErr)
	if err != nil {
		t.Errorf("can't do server.ListenAndServe() ")
	}

}

func TestServer_ListenAndServe_Negative(t *testing.T) {
	server, _ := New("tcp", "localhost", "50000", 1)

	nothingFunc := func(conn net.Conn) error {
		defer func() {
			_ = conn.Close()
		}()

		return nil
	}

	chErr := make(chan error, 1)

	err := server.ListenAndServe(nil, chErr)
	if err == nil {
		t.Errorf("no error for nil handler func")
	}

	err = server.ListenAndServe(nothingFunc, nil)
	if err == nil {
		t.Errorf("no error for nil error chan")
	}

}

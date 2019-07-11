package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/pkg/errors"
)

func ReadFromConnection(conn net.Conn, timeout time.Duration) (string, error) {

	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", errors.WithMessagef(err, "error while SetReadDeadline()")
	}

	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		if err, ok := err.(net.Error); !ok || !err.Timeout() {
			return "", errors.WithMessagef(err, "error while readFromConnection()")
		}
	}

	return string(buffer), nil
}

func WriteToConnection(conn net.Conn, timeout time.Duration, s string) error {
	err := conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil
	}
	_, err = fmt.Fprintf(conn, "%s\r\n", s)
	return errors.WithMessagef(err, "error while writeToConnection()")
}

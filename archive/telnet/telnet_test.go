package telnet

import (
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func Example() {

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handlers := CommandMap{
		"echo": func(connection net.Conn, msg string, args ...string) string {
			return msg
		},
	}

	server, err := NewTelnetServer("localhost", "8888", 1, 1, 5, handlers, log)
	if err != nil {
		os.Exit(1)
	}

	err = server.Start()
	if err != nil {
		os.Exit(2)
	}

	time.Sleep(time.Second * 100)

}

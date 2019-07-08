package server

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

type ConnType string

var (
	UDP ConnType = "udp"
	TCP ConnType = "tcp"
)

type Server struct {
	connType string
	host     string
	port     string
	l        net.Listener
	handler  func(conn net.Conn) error
}

func New(connType ConnType, host, port string) *Server {
	return &Server{
		connType: string(connType),
		host:     host,
		port:     port,
		l:        nil,
		handler:  nil,
	}
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

func (s *Server) RegHandleFunc(handler func(conn net.Conn) error) {
	s.handler = handler
}

func (s *Server) ListenAndServe(chErr chan<- error) error {
	err := s.Listen()
	if err == nil {
		go s.Start(chErr)
	}
	return err

}

func (s *Server) Listen() (err error) {
	s.l, err = net.Listen(s.connType, fmt.Sprintf("%s:%s", s.host, s.port))
	return
}

func (s *Server) Start(chErr chan<- error) {

	defer func() {
		if s.l != nil {
			err := s.l.Close()
			if err != nil {
				chErr <- errors.WithMessage(err, "problem in defer() func. net.Listener close()")
			}
		}
	}()

	for {
		// Listen for an incoming connection.
		conn, err := s.l.Accept()
		if err != nil {
			chErr <- errors.WithMessage(err, "problem accept new connection. net.Listener Accept()")
		}

		// Handle connections in a new goroutine.
		if s.handler != nil {
			go func() {
				err := s.handler(conn)
				if err != nil {
					chErr <- errors.WithMessagef(err, "err in handler func!")
				}
			}()
		} else {
			chErr <- errors.New("not redistricted handler func! handler func is nil")
		}
	}
}

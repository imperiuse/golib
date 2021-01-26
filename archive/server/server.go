package server

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

type (
	ConnType    string
	HandlerFunc = func(conn net.Conn) error
)

var (
	UDP ConnType = "udp"
	TCP ConnType = "tcp"
)

type Server struct {
	maxCntConnect int
	connType      string
	host          string
	port          string
	l             net.Listener
	handler       HandlerFunc
}

func New(connType ConnType, host, port string, maxCntConnect int) (*Server, error) {
	if host == "" || port == "" || connType == "" || (connType != TCP && connType != UDP) || maxCntConnect < 1 {
		return &Server{}, fmt.Errorf("bad input params! check params: connType:%s host:%s port:%s maxConn: %d",
			connType, host, port, maxCntConnect)
	}

	return &Server{
		maxCntConnect: maxCntConnect,
		connType:      string(connType),
		host:          host,
		port:          port,
		l:             nil,
		handler:       nil,
	}, nil
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.host, s.port)
}

func (s *Server) ListenAndServe(handler HandlerFunc, chErr chan<- error) error {
	if handler == nil {
		return errors.New("handler func is nil")
	}
	s.handler = handler

	err := s.Listen()
	if err == nil {
		go s.Start(chErr)
	}
	return err
}

func (s *Server) Listen() (err error) {
	s.l, err = net.Listen(s.connType, fmt.Sprintf("%s:%s", s.host, s.port))
	return err
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

	// create connect worker pool and connect chan
	chConn := make(chan net.Conn)
	for i := 0; i < s.maxCntConnect; i++ {
		go s.connectWorker(chConn, chErr)
	}

	for {
		// Listen for an incoming connection.
		conn, err := s.l.Accept()
		if err != nil {
			chErr <- errors.WithMessage(err, "problem accept new connection. net.Listener Accept()")
			continue
		}

		if conn == nil {
			chErr <- errors.WithMessage(err, "problem create new connection. net.Listener Accept()")
			continue
		}

		select {
		case chConn <- conn: // send conn to connect worker for process

		default: // all workers busy, reject connect
			chErr <- errors.New("All pool workers are busy")
			_ = conn.Close()
		}
	}
}

func (s *Server) connectWorker(chConn <-chan net.Conn, chErr chan<- error) {
	for {
		conn := <-chConn
		err := s.handler(conn)
		if err != nil {
			chErr <- err
		}
	}
}

package telnet

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/imperiuse/golib/archive/server"
	"github.com/pkg/errors"
)

type (
	ServerTelnet struct {
		server   *server.Server // base tcp server
		logger   *logrus.Logger
		timewait int        // timeout r/w
		timeout  int        // timeout close conn  //todo
		handlers CommandMap // telnet command handlers
	}
	Command        = string
	CommandMap     = map[Command]CommandHandler
	CommandHandler = func(connection net.Conn, msg string, args ...string) string
)

func NewTelnetServer(host, port string, maxConn, timeout, timewait int, handlers CommandMap, logger *logrus.Logger) (*ServerTelnet, error) {
	tcpServer, err := server.New("tcp", host, port, maxConn)
	if err != nil {
		return nil, errors.WithMessagef(err, "can't create new tcp server")
	}

	return &ServerTelnet{
		server:   tcpServer,
		logger:   logger,
		timewait: timewait,
		timeout:  timeout,
		handlers: handlers,
	}, nil
}

func (s *ServerTelnet) Start() error {
	chErr := make(chan error)
	go func() {
		for {
			err := <-chErr
			s.logger.Error(err)
		}
	}()

	err := s.server.ListenAndServe(s.TelnetMultiplexorHandler, chErr)
	if err != nil {
		return errors.WithMessagef(err, "can't ListenAndServe for telnet server")
	}

	s.logger.Infof("Telnet Server starts at %s", s.server.Addr())

	return nil
}

func (s *ServerTelnet) TelnetMultiplexorHandler(conn net.Conn) (err error) {
	var (
		request  string
		response string
	)

	defer func() {
		if err != nil { // чтобы не затирать входящую ошибку
			_ = conn.Close()
		} else {
			err = conn.Close()
		}
	}()

	oldCommand := ""
	oldArgs := []string{}
	reader := bufio.NewReader(conn)
	for {
		request, err = reader.ReadString(10)
		if err != nil {
			return err
		}
		sRequest := strings.Split(strings.Split(request, "\r\n")[0], " ")

		l := len(sRequest)
		if l > 0 {
			command := sRequest[0]
			args := make([]string, 0)

			// Повтор последней команды в стиле командной строки bash   //^[[A
			if len(command) == 3 &&
				[]byte(command)[0] == 27 &&
				[]byte(command)[1] == 91 &&
				[]byte(command)[2] == 65 {

				args = make([]string, len(oldArgs))
				copy(args, oldArgs)
				command = oldCommand
			}

			switch command {
			case "Q":
				fallthrough
			case "q":
				fallthrough
			case "exit":
				fallthrough
			case "quit":
				fallthrough
			case "Quit":
				fallthrough
			case "Exit":
				return
			default:
				if l > 1 {
					args = strings.Split(request, " ")[1:]
				}
				oldCommand = command
				oldArgs = make([]string, len(args))
				copy(oldArgs, args)
			}

			if handler, found := s.handlers[command]; found {
				response = handler(conn, command, args...)
			} else {
				unknownCommandText := fmt.Sprintf("Receive unknown command: %s", command)
				s.logger.Warning(unknownCommandText)
				response = unknownCommandText
			}
		}

		err = server.WriteToConnection(conn, time.Duration(s.timewait)*time.Second, response)
		if err != nil {
			return err
		}
	}
}

package telnet

import (
	"net"
	"strings"
	"time"

	"github.com/imperiuse/golib/server"
	"github.com/pkg/errors"
)

type (
	ServerTelnet struct {
		server   *server.Server // base tcp server
		timewait int            // timeout r/w
		timeout  int            // timeout close conn
		handlers CommandLists   // telnet command handlers
	}
	Command        = string
	CommandLists   = map[Command]CommandHandler
	CommandHandler = func(connection net.Conn, msg string, args ...string) string
)

func NewTelnetServer(host, port string, maxConn, timeout, timewait int, handlers CommandLists) (*ServerTelnet, error) {
	tcpServer, err := server.New("tcp", host, port, maxConn)
	if err != nil {
		return nil, errors.WithMessagef(err, "can't create new tcp server")
	}

	return &ServerTelnet{
		server:   tcpServer,
		timewait: timewait,
		timeout:  timeout,
		handlers: handlers,
	}, nil
}

func (s *ServerTelnet) Start() error {
	chErr := make(chan error, 0)
	go func() {
		for {
			err := <-chErr
			_ = err
			//w.logger.Error(err)
		}
	}()

	err := s.server.ListenAndServe(s.TelnetMultiplexorHandler, chErr)
	if err != nil {
		return errors.WithMessagef(err, "can't ListenAndServe for telnet server")
	}

	//s.logger.Infof("Telnet Server starts at %s", s.server.Addr())

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

	request, err = server.ReadFromConnection(conn, time.Duration(s.timewait)*time.Second)
	if err != nil {
		return err
	}

	sRequest := strings.Split(request, " ")
	l := len(sRequest)
	if l > 0 {
		command := sRequest[0]
		args := make([]string, 0)
		if l > 1 {
			args = strings.Split(request, " ")[1:]
		}

		if handler, found := s.handlers[command]; found {
			response = handler(conn, command, args...)
		}
	}

	err = server.WriteToConnection(conn, time.Duration(s.timewait)*time.Second, response)

	return err
}

//// CommandAnalyze - Функция обработки одного подключения
//// nolint
//func (server *ServerTelnet) CommandAnalyze(connection net.Conn, msgChan <-chan string, stopChan chan interface{}) {
//	// @param
//	// 	  connection   network.connection  - сетевое соединнение
//	// 	  msg     	 string              - анализируемое сообщеие (потенциальная команда)
//	var count int
//	var oldMsg string
//	for {
//		msg := <-msgChan
//		count++
//		Log.Debug("Telnet", "command_analyze()", fmt.Sprintf(" Receive Data %d: %s", count, msg))
//
//		// Повтор последней команды в стиле команжной строки bash
//		if len(msg) == 3 && []byte(msg)[0] == 27 && []byte(msg)[1] == 91 && []byte(msg)[2] == 65 { //^[[A
//			msg = oldMsg
//		} else {
//			oldMsg = msg
//		}
//
//		if msg == "Q" || msg == "q" {
//			msg = "exit"
//		}
//
//		noOneMatch := true
//		for _, command := range server.TCL {
//			if command.RegExp.MatchString(msg) {
//				noOneMatch = false
//				exit, err := command.Func(server, connection, msg)
//				if err != nil {
//					// BAD
//					Log.Error("Telnet", "command_analyze()", "ERR in CommandAnalyze F return")
//					stopChan <- new(interface{}) // Если что то не то закрываем ВСЕ!
//					return
//				}
//				if exit != nil {
//					Log.Info("Telnet", "command_analyze()", "Close CommandAnalyze()")
//					stopChan <- new(interface{}) // сигнал на закрытие всего
//					return
//				}
//				break // Ищем только самую первую команду в списке (учесть при формировании)
//			}
//		}
//
//		//  Unknown command
//		if noOneMatch {
//			Log.Info("Telnet", "command_analyze()", "Receive command: Unknown command!")
//			SafetyWrite(connection, fmt.Sprintf("Bad command send!  - %s ", msg))
//		}
//
//	}
//}
//
//// Функция обработки одного подключения
////  @param
////       connection  net.Conn  - сетевое соединение
////  @return
//func (server *ServerTelnet) handleConnection(connection net.Conn) {
//	defer func() { _ = connection.Close() }()
//	defer (*server.Stats).Dec("telnet_now_connect")
//	defer func() {
//		if r := recover(); r != nil {
//			Log.Error("Telnet", "handleConnection()", "Panic!", r)
//			_ = connection.Close()
//			(*server.Stats).Dec("telnet_connect")
//			(*server.Stats).Inc("telnet_panic_recover_handle_connection")
//		}
//	}()
//
//	t := time.After(time.Duration(server.Timeout) * time.Second) // Timeout после timeout sec секунд неактивности
//	// Каналы для связи двух go-рутин
//	msgChan := make(chan string)
//	stopChan := make(chan interface{})
//	// go-рутина анализатор сообщений (print command, switch and print result command or do smth...)
//	go server.CommandAnalyze(connection, msgChan, stopChan)
//	msgChan <- "help" // чтобы в начале вывелся списко команд
//
//	Log.Info("Telnet", "handleConnection()", fmt.Sprintf("Connection from %v established.", connection.RemoteAddr()))
//
//	buf := make([]byte, server.BufSize)
//
//	for {
//		_ = connection.SetReadDeadline(time.Now().Add(time.Second * 5))
//		n, err := connection.Read(buf)
//		if buf[0] == 0x04 {
//			err = io.EOF
//		}
//		if err != nil {
//			if err == io.EOF {
//				Log.Error("Telnet", "Connect close. EOF.", err)
//				msgChan <- "exit"           // команда для завершшения го-рутины обработки команд
//				time.Sleep(1 * time.Second) // чтобы дочка успела отработать и послать в канал stop сигнал стоп
//				//return ??
//			} else {
//				//Log.Debug("err read telnet", err)
//			}
//			goto next
//		}
//		if n == 0 {
//			//Log.Debug("Empty read")
//			goto next
//		}
//		// else no err and n >0
//		t = time.After(time.Duration(server.Timeout) * time.Second)
//		msgChan <- strings.TrimSpace(string(buf[0:n])) // передача команды для отработки
//		time.Sleep(250 * time.Millisecond)
//
//	next:
//		select {
//		case <-t: // timeout timeoutsec sec
//			Log.Info("Telnet", "handleConnection()",
//				fmt.Sprintf("Connection from %v closed. Timeout %v sec exist!.", connection.RemoteAddr(), server.Timeout))
//			return
//		case <-stopChan: // получена команда "Exit"
//			//c.Close()
//			Log.Info("Telnet", "handleConnection()",
//				fmt.Sprintf("Connection from %v closed. Exit сommand.", connection.RemoteAddr()))
//			return
//		default:
//			time.Sleep(150 * time.Millisecond)
//			break
//		}
//
//	}
//}

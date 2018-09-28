package telnet

import (
	"fmt"
	"github.com/imperiuse/golang_lib/colormap"
	"github.com/imperiuse/golang_lib/colors"
	gl "github.com/imperiuse/golang_lib/gologger"
	"github.com/imperiuse/golang_lib/safemap"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

// Структура описывающего TCP сервер для управления утилиты
type ServerTelnet struct {
	Port        uint             // Порт на кототором запущен TCP сервер (Telnet)
	Timeout     uint             // Таймаут после timeout sec секунд неактивности коннект закрывается - connect.Close()
	BufSize     uint             // Размер буфера для чтения
	LogFile     string           // Имя лог файла
	TCL         CommandList      // Список команд принимаемых и обрабатываемых
	Stats       *safemap.SafeMap // Статистика внешней отслеживаемой программы
	CommandChan chan Command     // Канал передачи команд управления и связи наверх
}

// Структура команды для передачи наверх (в функцию которую мониторим)
type Command struct {
	Name string
	Code int
	// VALUES
	int
	float64
	string
	ValueInterface interface{}
}

// Команда и ее обработка Telnet
type CommandTelnet struct {
	Name   string                                                                           // Наименование
	RegExp *regexp.Regexp                                                                   // RegExp для определения
	Func   func(server *ServerTelnet, connection net.Conn, msg string) (interface{}, error) // функция обработки команды
}

type CommandList []CommandTelnet

// Функция обработки одного подключения
// @param
// 	  connection   network.connection  - сетевое соединнение
// 	  msg     	 string              - анализируемое сообщеие (потенциальная команда)
func (server *ServerTelnet) CommandAnalyze(connection net.Conn, msgChan <-chan string, stopChan chan interface{}) {
	var count int = 0
	var oldMsg string
	for {
		msg := <-msgChan
		count++
		Log.Debug("Telnet", "command_analyze()", fmt.Sprintf(" Receive Data %d: %s", count, msg))

		// Повтор последней команды в стиле команжной строки bash
		if len(msg) == 3 && []byte(msg)[0] == 27 && []byte(msg)[1] == 91 && []byte(msg)[2] == 65 { //^[[A
			msg = oldMsg
		} else {
			oldMsg = msg
		}

		if msg == "Q" || msg == "q" {
			msg = "exit"
		}

		no_one_match := true
		for _, command := range server.TCL {
			if command.RegExp.MatchString(msg) {
				no_one_match = false
				exit, err := command.Func(server, connection, msg)
				if err != nil {
					// BAD
					Log.Error("Telnet", "command_analyze()", "ERR in CommandAnalyze F return")
					stopChan <- new(interface{}) // Если что то не то закрываем ВСЕ!
					return
				}
				if exit != nil {
					Log.Info("Telnet", "command_analyze()", "Close CommandAnalyze()")
					stopChan <- new(interface{}) // сигнал на закрытие всего
					return
				}
				break // Ищем только самую первую команду в списке (учесть при формировании)
			}
		}

		// UnKnown сommand
		if no_one_match {
			Log.Info("Telnet", "command_analyze()", "Receive command: Unknown command!")
			SafetyWrite(connection, fmt.Sprintf("Bad command send!  - %s ", msg))
		}

	}
}

// Функция обработки одного подключения
//  @param
//       connection  net.Conn  - сетевое соединение
//  @return
func (server *ServerTelnet) handleConnection(connection net.Conn) {
	defer connection.Close()
	defer (*server.Stats).Dec("telnet_now_connect")
	defer func() {
		if r := recover(); r != nil {
			Log.Error("Telnet", "handleConnection()", "Panic!", r)
			connection.Close()
			(*server.Stats).Dec("telnet_connect")
			(*server.Stats).Inc("telnet_panic_recover_handle_connection")
		}
	}()

	t := time.After(time.Duration(server.Timeout) * time.Second) // Timeout после timeout sec секунд неактивности
	// Каналы для связи двух go-рутин
	msgChan := make(chan string)
	stopChan := make(chan interface{})
	// go-рутина анализатор сообщений (print command, switch and print result command or do smth...)
	go server.CommandAnalyze(connection, msgChan, stopChan)
	msgChan <- "help" // чтобы в начале вывелся списко команд

	Log.Info("Telnet", "handleConnection()", fmt.Sprintf("Connection from %v established.", connection.RemoteAddr()))

	buf := make([]byte, server.BufSize)

	for {
		connection.SetReadDeadline(time.Now().Add(time.Second * 5))
		n, err := connection.Read(buf)
		if buf[0] == 0x04 {
			err = io.EOF
		}
		if err != nil {
			if err == io.EOF {
				Log.Error("Telnet", "Connect close. EOF.", err)
				msgChan <- "exit"           // команда для завершшения го-рутины обработки команд
				time.Sleep(1 * time.Second) // чтобы дочка успела отработать и послать в канал stop сигнал стоп
				//return ??
			} else {
				//Log.Debug("err read telnet", err)
			}
			goto next
		}
		if n == 0 {
			//Log.Debug("Empty read")
			goto next
		}
		// else no err and n >0
		t = time.After(time.Duration(server.Timeout) * time.Second)
		msgChan <- strings.TrimSpace(string(buf[0:n])) // передача команды для отработки
		time.Sleep(250 * time.Millisecond)

	next:
		select {
		case <-t: // timeout timeoutsec sec
			Log.Info("Telnet", "handleConnection()",
				fmt.Sprintf("Connection from %v closed. Timeout %v sec exist!.", connection.RemoteAddr(), server.Timeout))
			return
		case <-stopChan: // получена команда "Exit"
			//c.Close()
			Log.Info("Telnet", "handleConnection()",
				fmt.Sprintf("Connection from %v closed. Exit сommand.", connection.RemoteAddr()))
			return
		default:
			time.Sleep(150 * time.Millisecond)
			break
		}

	}
}

//Функция для форматирванной отправки текста по через TCP коннект (н-р на консоль подключившегося по Telnet)
//    @param
//        c     network.connect  - сетевое соединение
//        a...  string           - строки для отпарвляления в конце каждой строки добавляю x10 (/r) 0x13 (/n) перевод каретки и на новую строку)
//    @return
//
func SafetyWrite(c net.Conn, a ...string) {
	c.SetWriteDeadline(time.Now().Add(time.Second * 1))
	bytes := []byte("[Telnet]: ")
	for _, arg := range a {
		bytes = append(bytes, []byte(arg)...)
		bytes = append(bytes, byte(10), byte(13)) // //n //r
	}
	_, err := c.Write(bytes) // TODO надо ли анализировать сколько байт записалось или вообще это в цикл while поставить???
	if err != nil {
		Log.Error("Telnet", "SafetyWrite()", "Can't write data!", err, a)
	}
}

// Функция для логирования ошибки или статуса что ее нет.
func checkErrorFunc(err error, f string) {
	if err != nil {
		fmt.Println(fmt.Sprintf("[CheckErr %v]", f), colors.RED, "Error!", err, "\n", colors.RESET)
	} else {
		fmt.Println(fmt.Sprintf("[CheckErr %v]", f), colors.GREEN, "Successful!\n", colors.RESET)
	}
}

func recoveryFunc(f string) {
	if r := recover(); r != nil {
		Log.Error("Telnet: Recovery func for", f, r)
	}
}

var Log gl.Logger // Интерфейс для логирования
/*
* Основная функция Telnet
* На каждое открываемое соединение вызывается своя горутина, которая в свою очередь создает еще одну горутину для
* обработки получаемых данных первой. При получении кода EXIT/exit/Exit первая утилита посылает второй сообщение:
* на остановку, закрытие соединение и выход из goroutine-ы
 */

func (server *ServerTelnet) Run() {
	// Создание файла для логирования
	f, err := os.Create(server.LogFile + ".log")
	checkErrorFunc(err, "Create log file for Telnet")
	if err != nil {
		return
	}
	defer recoveryFunc("file_create") // Обработка возможной паники при создании файла
	defer f.Close()                   // Дефер для закрытия файла

	// Создание экземпляра Логгера для Логирования всех действий утилиты см. Пакет gologger/gologger.go
	Log = gl.NewLogger(os.Stdout, gl.ON_ALL, 1, 0, 0, "\t", colormap.CSMthemePicker("arseny"))

	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", server.Port))
	if err != nil {
		Log.Error("Telnet", "Run()", err)
		return
	}
	Log.Info("Telnet", "Run()", fmt.Sprintf("Start! Listening port: %v", server.Port))

	// INFINITE LOOP
	for {
		conn, err := ln.Accept() // Ждем подключения ( telnet IP port     // telnet 127.0.0.1 Port)
		if err != nil {
			Log.Error("Telnet", "Run()", err)
			continue
		}
		// Запускаем handler на подключение и возращаемся ждать нового подключения
		go server.handleConnection(conn)
		(*server.Stats).Inc("telnet_all_connect")
		(*server.Stats).Inc("telnet_now_connect")
	}
}

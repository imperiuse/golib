package wsclient

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	reconnect              = false
	maxCntReconnect        = 3
	maxCntLossPongResponse = 3
	outChanBuffer          = 10
)

type (
	Client struct {
		log logrus.FieldLogger

		addr, path string
		header     http.Header

		out     chan *outMsg // output msg to write ws connect
		handler HandlerFunc  // handler func to process input msg

		keepAlive        bool // создавать ли доп горутины для поддержания коннекта
		keepAliveTimeout time.Duration

		ctx    context.Context
		cancel context.CancelFunc

		reconnect bool

		sync.Mutex
		conn *gorilla.Conn
	}

	outMsg struct {
		msg        *gorilla.PreparedMessage
		errHandler ErrHandler
	}

	ErrHandler  = func(error)
	HandlerFunc = func(messageType int, p []byte, err error)
)

func New(scheme, host, port string, keepAlive bool, keepAliveTimeout time.Duration, log logrus.FieldLogger) *Client {
	return &Client{
		log:              log,
		addr:             fmt.Sprintf("%s://%s:%s", scheme, host, port),
		path:             "",
		header:           nil,
		out:              make(chan *outMsg, outChanBuffer),
		handler:          nil,
		keepAlive:        keepAlive,
		keepAliveTimeout: keepAliveTimeout,

		ctx:    nil,
		cancel: nil,

		reconnect: reconnect,
		Mutex:     sync.Mutex{},
	}
}

func (c *Client) ConnectAndListen(path string, header http.Header, handler HandlerFunc) error {
	err := c.connect(path, header)
	if err != nil {
		return err
	}

	if handler == nil {
		handler = func(messageType int, p []byte, err error) {
			if err != nil {
				return
			}
		}
	}
	c.handler = handler

	c.header = header

	go c.listen()

	return nil

}

func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Client) WriteMsg(msg *gorilla.PreparedMessage, errHandler ErrHandler) error {
	select {
	case c.out <- &outMsg{msg: msg, errHandler: errHandler}:
		return nil
	default:
		return errors.New("busy output channel")
	}
}

func NewPreparedTextMessage(msg []byte) (*gorilla.PreparedMessage, error) {
	return gorilla.NewPreparedMessage(gorilla.TextMessage, msg)
}

func (c *Client) connect(path string, header http.Header) error {
	var err error

	c.log.Infof("Attempt new WS connect to: %s/%s", c.addr, path)

	c.Lock()
	c.path = path
	c.conn, _, err = gorilla.DefaultDialer.Dial(fmt.Sprintf("%s/%s", c.addr, c.path), c.header)
	c.Unlock()

	if err != nil {
		return errors.WithMessagef(err, "can't dial by ws protocol to addr: %s/%s", c.addr, path)
	}
	return nil
}

func (c *Client) listen() {
	defer func() {
		_ = c.conn.Close()
	}()

	c.ctx, c.cancel = context.WithCancel(context.Background())

	if c.keepAlive {
		go c.startKeepAliveResponder()
	}

	// Listener
	go func() {
		for {
			// todo must be better solution this problem!! many dead conn reed  may be change signature of handler func
			func() {
				defer func() {
					err := recover()
					if err != nil {
						c.log.Error("Panic WS listen goroutine")
						c.Stop()
					}
				}()

				select {
				case <-c.ctx.Done():
					return
				default:
					c.handler(c.conn.ReadMessage())
				}
			}()
		}
	}()

	// Write and Ctx cancel
	for {
		select {
		// Send to server
		case msg := <-c.out:
			err := c.conn.WritePreparedMessage(msg.msg)
			if err != nil {
				c.log.WithError(err).Error("c.conn.WritePreparedMessage")
				if c.reconnect {
					c.Lock()
					for i := 0; i < maxCntReconnect; i++ {
						newConn, _, err := gorilla.DefaultDialer.Dial(fmt.Sprintf("%s/%s", c.addr, c.path), c.header)
						if err == nil {
							c.conn = newConn
							break
						}
						msg.errHandler(err)
						time.Sleep(time.Second)
					}
					c.Unlock()
				}
			}

		// Close and Exit
		case <-c.ctx.Done():
			c.Lock()
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
			if err != nil {
				c.log.WithError(err).Error("webSocket close error")
			}
			err = c.conn.Close()
			if err != nil {
				c.log.WithError(err).Error("result c.conn.Close()")
			}
			c.Unlock()
			<-time.After(time.Second)
			return
		}

	}
}

func (c *Client) startKeepAliveResponder() {
	ticker := time.NewTicker(c.keepAliveTimeout)
	defer ticker.Stop()

	lastResponse := time.Now()

	c.Lock()
	c.conn.SetPongHandler(func(msg string) error {
		c.log.Debug("Pong")
		lastResponse = time.Now()
		return nil
	})
	c.Unlock()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:

			deadline := time.Now().Add(10 * time.Second)

			c.Lock()
			err := c.conn.WriteControl(gorilla.PingMessage, []byte{}, deadline)
			c.Unlock()

			if err != nil {
				c.log.WithError(err).Error("problem write Ping to connect")
			} else {
				c.log.Debug("Ping")
			}

			<-ticker.C
			if time.Since(lastResponse) > maxCntLossPongResponse*c.keepAliveTimeout {
				c.log.Error("keepAliveTimeout exist")
				c.Stop()
				return
			}
		}
	}
}

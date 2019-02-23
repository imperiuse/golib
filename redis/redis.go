package redis

import (
	"fmt"
	"time"

	"github.com/imperiuse/golib/concat"

	"github.com/garyburd/redigo/redis"
	"github.com/imperiuse/golib/email"
	l "github.com/imperiuse/golib/logger"
)

// Redis - Bean struct for work with Redis
type Redis struct {
	Name string // Name DB (better uniq id in program)

	Host     string // host (url)
	Port     int    // tcp port number
	Password string // password
	DB       int    // db number

	RepeatFailDo       bool // repeat after fail command
	CountRepeatAttempt int  // try cnt repeat command  (RepeatFailDo == true)
	TimeRepeatAttempt  int  //  ! NanoSeconds ! time out repeat command (RepeatFailDo == true)

	MaxIdle     int           // max cnt idle connection
	MaxActive   int           // max cnt active
	IdleTimeout time.Duration // ! NanoSeconds !  max live time
	Wait        bool          // If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	MaxConnLifetime time.Duration // Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	TestOnBorrowTime time.Duration // ! NanoSeconds !  timeout for test on alive
	// TestOnBorrow() is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.

	Email  *email.MailBean // Mail Bean
	pool   *redis.Pool     // Pool connect к Redis
	Logger *l.LoggerI      // логгер
}

// InitNewPool - инициализировать внутренний пул подключений к Redis
func (r *Redis) InitNewPool() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error while init new redis pool. %v", r)
		}
	}()

	r.pool = &redis.Pool{
		MaxIdle:     r.MaxIdle,
		IdleTimeout: r.IdleTimeout,
		MaxActive:   r.MaxActive,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				concat.Strings(r.Host, fmt.Sprintf(":%v", r.Port)),
				redis.DialPassword(r.Password),
				redis.DialDatabase(r.DB))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < r.TestOnBorrowTime {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return nil
}

// GetName - get name obj Redis
func (r *Redis) GetName() string {
	return r.Name
}

// GetPool - get Redis pool obj
func (r *Redis) GetPool() *redis.Pool {
	return r.pool
}

func (r *Redis) doDefer(where string, com string, err error, args ...interface{}) {
	if rec := recover(); rec != nil {
		(*r.Logger).Error("[DEFER] Redis.doDefer()", where, r.Name, "PANIC!", rec)
		if err = r.Email.SendEmails(
			fmt.Sprintf("PANIC!\n%v\nErr:\n%+v\nSQL:\n%v\nWith args:\n%+v", where, rec, com, args)); err == nil {
			(*r.Logger).Error("pg.dbDefer()", where, r.Name, "Can't send email!", err)
		}
	}
}

// PingPongTest - Test work Redis query
func (r *Redis) PingPongTest() (err error) {
	conn := r.pool.Get()
	defer func() {
		if err = conn.Close(); err != nil {
			(*r.Logger).Error("Redis.PingPongTest()", r.Name, "Err while do conn.Close()", err)
		}
	}()

	var val string
	if val, err = redis.String(conn.Do("PING")); err != nil {
		(*r.Logger).Error("PingPongTest()", r.Name, "PING-PONG test Failed!", err)
		return
	}
	(*r.Logger).Info("PingPongTest()", r.Name, "PING-PONG test Successful Passed!", val)
	return nil
}

// Do - MAIN method for execute any Redis Command
func (r *Redis) Do(nameFuncWhoCall string, command string, args ...interface{}) (reply interface{}, err error) {
	defer r.doDefer(nameFuncWhoCall, command, err, args...)

	conn := r.pool.Get()
	defer func() {
		if err = conn.Close(); err != nil {
			(*r.Logger).Error("Redis.Do()", r.Name, "Err while do conn.Close()", err)
		}
	}()

	logging := nameFuncWhoCall != "" // если пустая строка в параметрах значит не логировать обращение к Redis
	for tryCnt := 0; tryCnt < r.CountRepeatAttempt; tryCnt++ {
		if logging {
			(*r.Logger).Debug(
				concat.Strings(nameFuncWhoCall, "--> Do()"),
				r.Name, fmt.Sprintf("Attemp execute Redis command: %v", tryCnt))
		}
		reply, err = conn.Do(command, args...)
		if err != nil {
			if logging {
				(*r.Logger).Log(l.RedisFail,
					concat.Strings(nameFuncWhoCall, "--> Do()"),
					concat.Strings(command, concat.Strings("", args[0].(string))),
					"Failed! Err:",
					err,
					"ARGS:",
					fmt.Sprintf("%v %v", args[0], args[1:]))
			}
			time.Sleep(time.Nanosecond * time.Duration(r.TimeRepeatAttempt))
			continue
		} else {
			if logging {
				(*r.Logger).Log(l.RedisOk,
					concat.Strings(nameFuncWhoCall, "--> Do()"),
					concat.Strings(command, concat.Strings("", args[0].(string))),
					"SUCCESSES!",
					"ARGS:",
					fmt.Sprintf("%v %v", args[0], args[1:]))
			}
			return
		}
	}
	err = fmt.Errorf("all try count end")
	(*r.Logger).Error(concat.Strings(nameFuncWhoCall, "--> Do()"), r.Name, "All try estimates! Panic!", err)
	panic(err)
}

package redis

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"golang_lib/gologger"
	"time"
)

// Структура настройки подключения к Redis
type Settings struct {
	Host     string // адрес сервера
	Port     int    // номер tcp порта которого слушаем
	Password string // пароль
	DB       int    // Номер БД
}

type RedisWorker struct {
	nameWorker         string           // имя воркера для логов и различия воркеров м/у собой
	pool               *redis.Pool      // Pool connect к Redis
	settings           *Settings        // настройки
	log                *gologger.Logger // логгер
	countRepeatAttempt int              // число попыток
	timeRepeatAttempt  int              // время между попытками
	repeatFailDo       bool             // пытаться повторить сделать запрос в редис
	emailSendingFunc   func(string)     // функция email уведомления о паниках
}

func CreateRedisWorker(NameWorker string, Pool *redis.Pool, settings *Settings, logger *gologger.Logger, emailSendingFunc func(string)) *RedisWorker {
	return &RedisWorker{NameWorker,
		Pool,
		settings,
		logger,
		5,
		10,
		true,
		emailSendingFunc,
	}
}

func CreateNewPool(MaxIdle int, IdleTimeout time.Duration, MaxActive int, TestOnBorrowTime time.Duration, settings Settings) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     MaxIdle,
		IdleTimeout: IdleTimeout,
		MaxActive:   MaxActive,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				settings.Host+":"+fmt.Sprintf("%v", settings.Port),
				redis.DialPassword(settings.Password),
				redis.DialDatabase(settings.DB))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < TestOnBorrowTime {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (rw *RedisWorker) redisWorkerDefer(funcName string, f_recovery func()) {
	if r := recover(); r != nil {
		(*rw.log).Error("Defer! For "+funcName, fmt.Sprintf("RedisWorker:%v", rw.nameWorker), "PANIC!", r)
		f_recovery()
	}
}

// Test work Redis query
func (rw *RedisWorker) PingPongTest() error {
	conn := rw.pool.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("PING"))
	if err != nil {
		(*rw.log).Error("createRedisWorker()",
			fmt.Sprintf("RedisWorker:%v", rw.nameWorker),
			"PING-PONG Test Failed!",
			err)
		return err
	} else {
		(*rw.log).Info("createRedisWorker()",
			fmt.Sprintf("RedisWorker:%v", rw.nameWorker),
			"PING-PONG Test Passed! Good!", val)
		return nil
	}
}

func (rw *RedisWorker) MyName() string {
	return rw.nameWorker
}

func (rw *RedisWorker) Do(nameFunc string, commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := rw.pool.Get()
	defer conn.Close()

	defer rw.redisWorkerDefer(nameFunc+"-->"+"DO()", func() {
		(*rw.log).Error("Recover() <-- Do()", fmt.Sprintf("RedisWorker:%v", rw.nameWorker), "Panic was!")
		if rw.emailSendingFunc != nil {
			rw.emailSendingFunc(fmt.Sprintf("%v --> %v", nameFunc, "DO() Panic was!"))
		}
		// stats panic cnt inc // todo
	})

	logging := nameFunc != "" // если пустая строка в параметрах значит не логирвоать обращение к Redis
	for try_cnt := 0; try_cnt < rw.countRepeatAttempt; try_cnt++ {
		if logging {
			(*rw.log).Debug(
				nameFunc+"-->"+"Do()", fmt.Sprintf("RedisWorker:%v", rw.nameWorker),
				fmt.Sprintf("Attemp execute Redis command: %d", try_cnt),
			)
		}
		reply, err = conn.Do(commandName, args...)
		if err != nil {
			if logging {
				(*rw.log).Log(gologger.REDIS_FAIL, commandName+" "+args[0].(string), nameFunc+"-->"+"Do()", "Failed! Err:", err, "ARGS:", fmt.Sprintf("%v %x", args[0], args[1:]))
			}
			continue
		} else {
			if logging {
				(*rw.log).Log(gologger.REDIS_OK, commandName+" "+args[0].(string), nameFunc+"-->"+"Do()", "SUCCESSES!", "ARGS:", fmt.Sprintf("%v %x", args[0], args[1:]))
			}
			return
		}
	}
	err = errors.New("try count end")
	(*rw.log).Error(nameFunc+"-->"+"Do()", fmt.Sprintf("RedisWorker:%v", rw.nameWorker), "All try estimates! Panic!", err)
	panic(err)
}

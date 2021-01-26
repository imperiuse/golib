package redis

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"gitlab.esta.spb.ru/arseny/golib/email"
)

const (
	defaultNetwork          = "tcp"
	defaultHost             = "localhost"
	defaultPort             = "6379"
	defaultMaxIdleConn      = 50
	defaultMaxActive        = 100
	defaultIdleTimeout      = time.Second * 60
	defaultMaxConnLifetime  = time.Second * 300
	defaultTestOnBorrowTime = time.Second * 60
)

type (
	Redis struct {
		Name string // Name DB (better uniq id in program)

		cfg  Config
		lg   log.FieldLogger
		pool *redis.Pool

		email *email.MailBean // Mail Bean
	}

	Config struct {
		// Do `FLUSHALL` redis command at the beginning
		FlushAll bool `mapstructure:"flush"`

		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`

		// Redigo param's
		// Maximum number of idle connections in the pool.
		MaxIdle int `mapstructure:"max_idle"`

		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool.
		MaxActive int `mapstructure:"max_Active"`

		// Close connections after remaining idle for this duration. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		IdleTimeout time.Duration `mapstructure:"IdleTimeout"`

		// If Wait is true and the pool is at the MaxActive limit, then Get() waits
		// for a connection to be returned to the pool before returning.
		Wait bool `mapstructure:"wait"`

		// Close connections older than this duration. If the value is zero, then
		// the pool does not close connections based on age.
		MaxConnLifetime time.Duration `mapstructure:"max_conn_life_time"`

		// TestOnBorrow() is an optional application supplied function for checking
		// the health of an idle connection before the connection is used again by
		// the application. Argument t is the time that the connection was returned
		// to the pool. If the function returns an error, then the connection is
		// closed.
		TestOnBorrowTime time.Duration `mapstructure:"test_on_borrow_time"`
	}
)

type ConfigParams = map[string]interface{}

func NewRedisDB(config ConfigParams, lg log.FieldLogger) (*Redis, error) {

	r := Redis{
		cfg:  Config{},
		lg:   lg,
		pool: nil,
	}

	err := mapstructure.Decode(config, &r.cfg)
	if err != nil {
		return nil, err
	}

	if r.cfg.Host == "" {
		r.cfg.Host = defaultHost
	}

	if r.cfg.Port == "" {
		r.cfg.Port = defaultPort
	}

	if r.cfg.MaxIdle == 0 {
		r.cfg.MaxIdle = defaultMaxIdleConn
	}

	if r.cfg.IdleTimeout == 0 {
		r.cfg.IdleTimeout = defaultIdleTimeout
	}

	if r.cfg.MaxActive == 0 {
		r.cfg.MaxActive = defaultMaxActive
	}

	if r.cfg.MaxConnLifetime == 0 {
		r.cfg.MaxConnLifetime = defaultMaxConnLifetime
	}

	if r.cfg.TestOnBorrowTime == 0 {
		r.cfg.TestOnBorrowTime = defaultTestOnBorrowTime
	}

	r.pool = &redis.Pool{
		MaxIdle:         r.cfg.MaxIdle,
		IdleTimeout:     r.cfg.IdleTimeout,
		MaxActive:       r.cfg.MaxActive,
		MaxConnLifetime: r.cfg.MaxConnLifetime,
		Wait:            r.cfg.Wait,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				defaultNetwork,
				net.JoinHostPort(r.cfg.Host, r.cfg.Port),
				redis.DialPassword(r.cfg.Password),
				redis.DialDatabase(r.cfg.DB))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < r.cfg.TestOnBorrowTime {
				return nil
			}
			_, err := c.Do("PING")

			if err != nil {
				r.lg.WithError(err).Error("ping to Redis problem")
			}

			return err
		},
	}

	if err := r.PingPongTest(); err != nil {
		return nil, err
	}

	if r.cfg.FlushAll {
		if err := r.FlushAll(); err != nil {
			return nil, fmt.Errorf("redis FLUSHALL command problem: %w", err)
		}
		r.lg.Info("Successfully FlushAll Redis DB")
	}

	r.lg.Info("Create new Redis connections pool")

	return &r, nil
}

func (r *Redis) Shutdown() {
	err := r.pool.Close()
	if err != nil {
		r.lg.WithError(err).Error("Failed closed Redis conn pool")
		return
	}
	r.lg.Info("Successfully closed Redis conn pool")
}

func (r *Redis) closeConnErrLog(conn redis.Conn) {
	if err := conn.Close(); err != nil {
		r.lg.WithError(err).Error("close redis conn problem")
	}
}

func (r *Redis) PingPongTest() error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		r.lg.WithField("Command", "PingPongTest").Error("PING-PONG test Failed!")
		return err
	}

	r.lg.WithField("Command", "PingPongTest").Info("Ping-Pong test Successful Passed!")
	return nil
}

func (r *Redis) HSet(name, field, value string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	res, err := conn.Do("HSET", name, field, value)

	if err != nil {
		return err
	}

	r.lg.Debugf("hset: returncode is %v", res)
	return nil
}

func (r *Redis) HGet(key, field string) ([]byte, error) {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	value, err := redis.Bytes(conn.Do("HGET", key, field))

	if err != nil && !errors.Is(err, redis.ErrNil) {
		return []byte{}, err
	}

	return value, nil
}

func (r *Redis) HDel(key string, fields ...string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	values := make([]interface{}, len(fields)+1)
	values[0] = key
	for i, val := range fields {
		values[i+1] = val
	}

	res, err := conn.Do("HDEL", values...)

	if err != nil {
		return err
	}

	r.lg.Debugf("hdel: result is %v", res)
	return nil
}

func (r *Redis) RPush(key string, args ...string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	values := make([]interface{}, len(args)+1)
	values[0] = key
	for i, val := range args {
		values[i+1] = val
	}

	res, err := conn.Do("RPUSH", values...)

	if err != nil {
		return err
	}

	r.lg.Debugf("rpush: returncode is %v", res)
	return nil
}

func (r *Redis) LIndex(key string, index int64) (string, error) {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	value, err := redis.String(conn.Do("LINDEX", key, index))

	if err != nil && !errors.Is(err, redis.ErrNil) {
		return "", err
	}

	return value, nil
}

func (r *Redis) SAdd(key string, members ...string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	return errors.New("SADD is not implemented for redis")
}

func (r *Redis) LPop(key string) ([]byte, error) {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	res, err := redis.Bytes(conn.Do("LPOP", key))

	if err != nil {
		return []byte{}, err
	}

	return res, nil
}

func (r *Redis) LRem(key string, count int64, val string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := conn.Do("LREM", key, count, val)
	return err
}

func (r *Redis) ZAdd(key string, member string, score int64) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := conn.Do("ZADD", key, float64(score), member)
	return err
}

func (r *Redis) ZRem(key string, member string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := conn.Do("ZREM", key, member)
	return err
}

func (r *Redis) ZRemByScore(key string, score string) error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := conn.Do("ZREMRANGEBYSCORE", key, score, score)
	return err
}

func (r *Redis) ZRange(key string, start int64, stop int64) ([]string, error) {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	return redis.Strings(conn.Do("ZRANGE", key, start, stop))
}

func (r *Redis) FlushAll() error {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	_, err := conn.Do("FLUSHALL")
	return err
}

func (r *Redis) GetName() string {
	return r.Name
}

func (r *Redis) GetPool() *redis.Pool {
	return r.pool
}

func (r *Redis) Do(lg log.FieldLogger, command string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer r.closeConnErrLog(conn)

	if lg == nil {
		lg = r.lg
	}

	res, err := conn.Do(command, args...)
	if err != nil {
		lg.WithError(err).Error("problem Do command: `%s` with args: `%v`", command, args)
	} else {
		lg.Debug("Result of command: `%s %s`  -  %v", command, args, res)
	}

	return res, err
}

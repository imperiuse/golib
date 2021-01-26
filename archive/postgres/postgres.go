package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/imperiuse/golib/archive/concat"

	l "github.com/imperiuse/golib/archive/logger"
	"github.com/imperiuse/golib/email"

	"github.com/jmoiron/sqlx"
	//nolint
	_ "github.com/lib/pq"
)

// SSL mod
const (
	DisableSSL    = "disable"
	RequireSSL    = "require"
	VerifyCASSL   = "verify-ca"
	VerifyFullSSL = "verify-full"
)

// PgDB - Bean for work with Postgres DB
type PgDB struct {
	Name string // Name DB (better uniq id in program)

	URL    string // Domain name (localhost - default)
	Host   string // Hostname domain (IP)
	Port   int    // Port Db (Postgres 5432)
	DbName string // Db name (main)

	SSL string // SSL mod (disable/enable)  @see https://godoc.org/github.com/lib/pq
	// * disable - No SSL * require - Always SSL (skip verification) * verify-ca - Always SSL * verify-full - Always SSL
	SSLCert     string // sslcert Cert file location. The file must contain PEM encoded data.
	SSLKey      string // sslkey Key file location. The file must contain PEM encoded data.
	SSLRootCert string // sslrootcert The location of the root certificate file. The file

	User string // The user to sign in as
	Pass string // The user's password

	CntAttemptRequest  int  // Cnt attempts connect to DB
	TimeAttemptRequest int  // Time between attempts  ! SECONDS !
	RepeatRequest      bool // Cnt try repeat execute SQL request to DB
	ConnMaxLifetime    int  // time in Nanosecond
	MaxIdleConns       int  // max idle connections
	MaxOpenConns       int  // max open connections

	Email  *email.MailBean // Email Bean for send error info
	Logger l.LoggerI       // Pointer to Logger interface
	db     *sqlx.DB        // Pool connection to DB (return by sql.Open("postgres", ".....db_settings))
}

// IPgDB - public interface describes PgDB
type IPgDB interface {
	GetDB() *sqlx.DB
	ExecuteQuery(string, string, ...interface{}) (*sql.Rows, error)
	ExecuteRowAffected(string, string, ...interface{}) (int64, error)
	ExecuteQueryX(string, string, ...interface{}) (*sqlx.Rows, error)
	Select(string, interface{}, string, ...interface{}) error
	Get(string, interface{}, string, ...interface{}) error
	NamedExec(string, string, map[string]interface{}) (sql.Result, error)
	NamedQuery(string, string, interface{}) (*sqlx.Rows, error)
}

// GetDB - get current DB connect
func (pg *PgDB) GetDB() *sqlx.DB {
	return pg.db
}

// GetName - get Name obj DB
func (pg *PgDB) GetName() string {
	return pg.Name
}

// ConfigString - config connect DB
func (pg *PgDB) ConfigString() (config string) {
	switch pg.SSL {
	case VerifyFullSSL:
		fallthrough
	case VerifyCASSL:
		fallthrough
	case RequireSSL:
		config = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s "+
			"sslmod=%s sslcert=%s sslkey=%s sslrootcert=%s",
			pg.Host, pg.Port, pg.DbName, pg.User, pg.Pass, pg.SSL, pg.SSLCert, pg.SSLKey, pg.SSLRootCert)
	case DisableSSL:
		fallthrough
	default:
		config = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmod=%s ",
			pg.Host, pg.Port, pg.DbName, pg.User, pg.Pass, pg.SSL)
	}
	return
}

// Connect - Создание пула коннекшенов к БД
func (pg *PgDB) Connect() (err error) {
	// Use connect for "true" connect check to DB
	if pg.db, err = sqlx.Connect("postgres", pg.ConfigString()); err != nil {
		pg.Logger.Error("PgDB.Connect()", pg.Name, "Can't connect to DB server!",
			pg.ConfigString(), err)
	}
	if pg.ConnMaxLifetime > 0 {
		pg.db.SetConnMaxLifetime(time.Duration(pg.ConnMaxLifetime))
	}
	if pg.MaxIdleConns > 0 {
		pg.db.SetMaxIdleConns(pg.MaxIdleConns)
	}
	if pg.MaxOpenConns > 0 {
		pg.db.SetMaxOpenConns(pg.MaxOpenConns)
	}
	return err
}

// Close - Закрытие соединения
func (pg *PgDB) Close() {
	if err := pg.db.Close(); err != nil {
		pg.Logger.Error("PgDB.close()", pg.Name, "Can't close DB connection!", err)
	}
	pg.Logger.Info("PgDB.close()", pg.Name,
		concat.Strings("Connection to database", pg.Host, ":", pg.DbName, "successful close()"))
}

func (pg *PgDB) executeDefer(callBy string, query string, err error, args ...interface{}) {
	if r := recover(); r != nil {
		pg.Logger.Error(callBy, pg.Name, "PANIC!", r)
		if errIn := pg.Email.SendEmails(
			fmt.Sprintf("PANIC!\n%v\nErr:\n%+v\nSQL:\n%v\nWith args:\n%+v", callBy, r, query, args)); err == nil {
			pg.Logger.Error(callBy, pg.Name, "Can't send email!", errIn)
		}
	}
}

// ExecuteQuery - Функция обертка над execute. Запросы с ожиданием данных от БД. (SELECT и т.д. возращающие значения)
func (pg *PgDB) ExecuteQuery(callBy string, query string, args ...interface{}) (rows *sql.Rows, err error) {
	callBy = concat.Strings(callBy, " postgres.ExecuteQuery()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, args...)

	for cnt := 0; cnt < pg.CntAttemptRequest; cnt++ {
		pg.Logger.Debug(callBy, pg.Name, concat.Strings("Attemp execute query: ", strconv.Itoa(cnt)))
		rows, err = pg.db.Query(query, args...)
		if err != nil {
			pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED",
				err,
				query,
				"ARGS:", args)
		} else {
			pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
				query,
				"ARGS:", args)
			return rows, nil
		}
		time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
	}
	err = fmt.Errorf("all try estimates! %v", err)
	return nil, err
}

// ExecuteRowAffected - Функция обертка над Execute. Запрос без ожидания данных, с ожиданием кол-ва затронутых строк.
func (pg *PgDB) ExecuteRowAffected(callBy string, query string, args ...interface{}) (rowAffected int64, err error) {
	callBy = concat.Strings(callBy, " postgres.ExecuteRowAffected()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, args...)

	for cnt := 0; cnt < pg.CntAttemptRequest; cnt++ {
		var results sql.Result
		results, err = pg.db.Exec(query, args...)
		if err != nil {
			pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED",
				err,
				query,
				"ARGS:", args)
		} else {
			pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
				query,
				"ARGS:", args)
			if rowAffected, err = results.RowsAffected(); err != nil {
				pg.Logger.Error("Err while rows affected:", err)
			}
			return rowAffected, err
		}
		time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
	}
	err = fmt.Errorf("all try estimates! %v", err)
	return rowAffected, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////   Расширение SQLX              //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ExecuteQueryX - Функция обертка над QueryX (sqlX). Запросы с ожиданием данных от БД.
func (pg *PgDB) ExecuteQueryX(callBy string, query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	callBy = concat.Strings(callBy, " --> postgres.ExecuteQueryX()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, args...)

	for cnt := 0; cnt < pg.CntAttemptRequest; cnt++ {
		pg.Logger.Debug(callBy, pg.Name, concat.Strings("Attemp execute query: ", strconv.Itoa(cnt)))
		rows, err = pg.db.Queryx(query, args...)
		if err != nil {
			pg.Logger.Log(l.DbFail, query, pg.Name, "SQL FAILED",
				err,
				query,
				"ARGS:", args)
		} else {
			pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
				query,
				"ARGS:", args)
			return rows, nil
		}
		time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
	}
	err = fmt.Errorf("all try estimates! %v", err)
	return nil, err
}

// Select - syntax sugar of `SELECT ... ` method;
// @param
//     callBy       string      - кто вызвал, важно для логирования, чтобы не вызывать runtime.Caller()
//     dest         interface   - интерфейс указатель куда запишем данные
//     query        string      - строка SQL запрос
//     args...      interface{} - аргументы
//  @return
//                  error       - есть ли ошибка в запросе
func (pg *PgDB) Select(callBy string, dest interface{}, query string, args ...interface{}) (err error) {
	callBy = concat.Strings(callBy, " --> postgres.Select()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, args...)

	// NOTE  // if you have null fields and use SELECT *, you must use sql.Null* in your struct
	if err = pg.db.Select(dest, query, args...); err != nil {
		pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED", err,
			query,
			"ARGS:", args)
	} else {
		pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
			query,
			"ARGS:", args)
	}
	return
}

// Get - syntax sugar of `SELECT ... LIMIT 1` method;
// @param
//     callBy       string      - кто вызвал, важно для логирования, чтобы не вызывать runtime.Caller()
//     dest         interface   - интерфейс указатель куда запишем данные
//     query        string      - строка SQL запрос
//     args...      interface{} - аргументы
//  @return
//                  error       - есть ли ошибка в запросе
func (pg *PgDB) Get(callBy string, dest interface{}, query string, args ...interface{}) (err error) {
	callBy = concat.Strings(callBy, " --> postgres.Get()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, args...)

	if err = pg.db.Get(dest, query, args...); err != nil {
		pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED", err,
			query,
			"ARGS:", args)
	} else {
		pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
			query,
			"ARGS:", args)
	}
	return
}

// NamedExec - syntax sugar of sql.NamedExec `INSERT INTO person (first_name,last_name,email) VALUES (:first,:last,:email)`,;
// 													map[string]interface{}{
// 													           "first": "Bin",
// 													           "last": "Smuth",
// 													           "email": "bensmith@allblacks.nz",
// @param
//     callBy       string                 - кто вызвал, важно для логирования, чтобы не вызывать runtime.Caller()
//     query        string                 - строка SQL запрос
//     nameArgs     map[string]interface{} - map с именнованными аргументами
//  @return
//                  sql.Result  - результаты запроса
//                  error       - есть ли ошибка в запросе
func (pg *PgDB) NamedExec(callBy string, query string, nameArgs map[string]interface{}) (result sql.Result, err error) {
	callBy = concat.Strings(callBy, " --> postgres.NamedExec()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, nameArgs)

	if result, err = pg.db.NamedExec(query, nameArgs); err != nil {
		pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED", err,
			query,
			"ARGS:", nameArgs)
	} else {
		pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
			query,
			"ARGS:", nameArgs)
	}
	return
}

// NamedQuery - syntax sugar of sql.NamedQuery `SELECT * FROM person WHERE first_name=:first_name`, jason`,;
// @param
//     callBy       string      - кто вызвал, важно для логирования, чтобы не вызывать runtime.Caller()
//     query        string      - строка SQL запрос
//     data         interface{} - map с именнованными аргументами
//  @return
//                  *sqlx.Rows  - результаты зароса строки
//                  error       - есть ли ошибка в запросе
func (pg *PgDB) NamedQuery(callBy string, query string, data interface{}) (rows *sqlx.Rows, err error) {
	callBy = concat.Strings(callBy, " --> postgres.NamedQuery()")
	defer pg.executeDefer(concat.Strings(callBy, " [DEFER]"), query, err, data)

	if rows, err = pg.db.NamedQuery(query, data); err != nil {
		pg.Logger.Log(l.DbFail, callBy, pg.Name, "SQL FAILED", err,
			query,
			"ARGS:", data)
	} else {
		pg.Logger.Log(l.DbOk, callBy, pg.Name, "SQL SUCCESS",
			query,
			"ARGS:", data)
	}
	return
}

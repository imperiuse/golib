package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/imperiuse/golib/email"
	l "github.com/imperiuse/golib/logger"

	"github.com/jmoiron/sqlx"
	//_ "github.com/lib/pq"
)

// PgDB - Bean for work with Postgres DB
type PgDB struct {
	Name string // Name DB (uniq id in program)

	URL    string // Domain name (localhost - default)
	Host   string // Hostname domain (IP)
	Port   string // Port Db (Postgres 5432)
	DbName string // Db name (main)
	SSL    string // SSL mod (disable/enable)
	User   string // User name
	Pass   string // Password user

	CntAttemptRequest  int  // Cnt attempts connect to DB
	TimeAttemptRequest int  // Time between attempts
	RepeatRequest      bool // Cnt try repeat execute SQL request to DB

	Email  *email.MailBean // Email Bean for send error info
	logger *l.Logger       // Pointer to logger interface
	db     *sqlx.DB        // Pool connection to DB (return by sql.Open("postgres", ".....db_settings))
}

// GetDB - get current DB connect
func (pg *PgDB) GetDB() *sqlx.DB {
	return pg.db
}

// ConfigString - config connect DB
func (pg *PgDB) ConfigString() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s user=%s password=%s",
		pg.Host, pg.Port, pg.DbName, pg.SSL, pg.User, pg.Pass)
}

// Connect - Создание пула коннекшенов к БД
func (pg *PgDB) Connect() (err error) {
	if pg.db, err = sqlx.Open("postgres", pg.ConfigString()); err != nil {
		(*pg.logger).Error("PgDB.Connect()", pg.Name, "Can't open (get handle to database) to DB server!",
			pg.ConfigString(), err)
	}
	if err = pg.db.Ping(); err != nil {
		(*pg.logger).Error("PgDB.Connect()", pg.Name, "Can't open connect (can't Ping) to DB server!",
			pg.ConfigString(), err)
	}
	return err
}

// Close - Закрытие соединения
func (pg *PgDB) Close() {
	if err := pg.db.Close(); err != nil {
		(*pg.logger).Error("PgDB.close()", pg.Name, "Can't close DB connection!", err)
	}
	(*pg.logger).Info("PgDB.close()", pg.Name,
		fmt.Sprintf("Connection to database %v:%v successfull close()", pg.Host, pg.DbName))
}

func (pg *PgDB) executeDefer(where string, query string, err error, args ...interface{}) {
	if r := recover(); r != nil {
		(*pg.logger).Error("Defer! For "+where, pg.Name, "PANIC!", r)
		if pg.Email != nil {
			if err = pg.Email.SendEmailByDefaultTemplate(
				fmt.Sprintf("PANIC!\n%v\nErr:\n%+v\nSQL:\n%v\nWith args:\n%+v", where, r, query, args)); err == nil {
				(*pg.logger).Error("pg.dbDefer()", where, pg.Name, "Can't send email!", err)
			}
		}
	}
}

// ExecuteQuery - Функция обертка над execute. Запросы с ожиданием данных от БД. (SELECT и т.д. возращающие значения)
func (pg *PgDB) ExecuteQuery(nameFunc string, query string, args ...interface{}) (rows *sql.Rows, err error) {
	defer pg.executeDefer(nameFunc+"-->"+"ExecuteQuery()", query, err, args...)
	return pg.execute(true, nameFunc, query, args...), err
}

// Execute - Функция обертка над execute. Запросы без ожидания данных от БД. (Update и т.д. не возращающие значения)
func (pg *PgDB) Execute(nameFunc string, query string, args ...interface{}) (err error) {
	defer pg.executeDefer(nameFunc+"-->"+"Execute()", query, err, args...)
	_ = pg.execute(false, nameFunc, query, args...)
	return err
}

// Функция выполнения запроса query
// @param
//     return_flag  bool        -  флаг типа запроса
//     name_func    string      -  имя вызывающей функции для логирования
//     SQL          string      -  строка SQL запрос
//     args...      interface{} - аргументы
// @return
//     row          sql.Rows -  результат запроса, данные от БД
func (pg *PgDB) execute(returnValue bool, nameFunc string, SQL string, args ...interface{}) (row *sql.Rows) {
	var err error
	// Проверка коннекта к БД
	if err = pg.db.Ping(); err != nil {
		(*pg.logger).Error(nameFunc+"-->"+"execute()", "DbWorker", "Can't open connect (can't Ping) to Db server!", err)
		err = errors.New("no connect")
		panic(err)
	}
	for cnt := 0; cnt < pg.CntAttemptRequest; cnt++ {
		(*pg.logger).Debug(nameFunc+"-->"+"execute()", "DbWorker", fmt.Sprintf("Attemp execute SQL: %d", cnt))
		if returnValue { // TRUE == Execute_Query
			row, err = pg.db.Query(SQL, args...)
			if err != nil {
				(*pg.logger).Log(l.DbFail, SQL, nameFunc+"-->"+"Execute_Query()", "Failed! Err:", err, "ARGS:", args)
			} else {
				(*pg.logger).Log(l.DbOk, SQL, nameFunc+"-->"+"Execute_Query()", "SUCCESSES!", "ARGS:", args)
				return row
			}
			time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
		} else { // FALSE == Execute
			var results sql.Result
			results, err = pg.db.Exec(SQL, args...)
			if err != nil {
				(*pg.logger).Log(l.DbFail, SQL, nameFunc+"-->"+"Execute()", "Failed! Err:", err, "ARGS:", args)
			} else {
				(*pg.logger).Log(l.DbOk, SQL, nameFunc+"-->"+"Execute()", "SUCCESSES!", "ARGS:", args)
				if rA, err1 := results.RowsAffected(); err1 == nil {
					(*pg.logger).Info("Rows affected:", rA)
				}
				return nil
			}
		}
	}
	(*pg.logger).Error(nameFunc+"()-->"+"execute()", "DbWorker", "All try estimates! Panic!", err)
	panic(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////   Расширение SQLX              //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ExecuteQueryX - Функция обертка над QueryX (sqlX). Запросы с ожиданием данных от БД.
func (pg *PgDB) ExecuteQueryX(nameFunc string, query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	defer pg.executeDefer(nameFunc+"-->"+"ExecuteQueryX()", query, err, args...)
	return pg.executeX(nameFunc, query, args...), err
}

// Функция выполнения запроса queryX
// @param
//     name_func    string      -  имя вызывающей функции для логирования
//     SQL          string      -  строка SQL запрос
//     args...      interface{} - аргументы
//  @return
//     row          *sqlx.Rows
func (pg *PgDB) executeX(nameFunc string, SQL string, args ...interface{}) (row *sqlx.Rows) {
	var err error
	// Проверка коннекта к БД
	if err = pg.db.Ping(); err != nil {
		(*pg.logger).Error(nameFunc+"-->"+"executeX()", "DbWorker", "Can't open connect (can't Ping) to Db server!", err)
		err = errors.New("no connect")
		panic(err)
	}
	for cnt := 0; cnt < pg.CntAttemptRequest; cnt++ {
		(*pg.logger).Debug(nameFunc+"-->"+"executeX()", "DbWorker", fmt.Sprintf("Attemp execute SQL: %d", cnt))
		row, err = pg.db.Queryx(SQL, args...)
		if err != nil {
			(*pg.logger).Log(l.DbFail, SQL, nameFunc+"-->"+"ExecuteX QueryX()", "Failed! Err:", err, "ARGS:", args)
		} else {
			(*pg.logger).Log(l.DbOk, SQL, nameFunc+"-->"+"ExecuteX QueryX()", "SUCCESSES!", "ARGS:", args)
			return row
		}
		time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
	}
	(*pg.logger).Error(nameFunc+"()-->"+"executeX()", "DbWorker", "All try estimates! Panic!", err)
	panic(err)
}

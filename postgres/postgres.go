package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/imperiuse/golib/email"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PgDB - Bean for work with Postgres DB
type PgDB struct {
	Url      string // Domain name (www..)
	Host     string // Hostname domain (IP)
	Port     string // Port Db (Postgres 5432)
	Database string // Db name (main)
	SSL      string // SSL mod (disable/enable)
	User     string // User name
	Pass     string // Password user

	CntAttemptRequest  uint // Cnt attempts connect to DB
	TimeAttemptRequest uint // Time between attempts
	RepeatRequest      bool // Cnt try repeat execute SQL request to DB

	Email  *email.MailBean // Email Bean for send error info
	logger *logger.Logger  // Pointer to logger interface
	db     *sqlx.DB        // Pool connection to DB (return by sql.Open("postgres", ".....db_settings))
}

func (dbS *PgDB) String() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s "+
		"sslmode=%s user=%s password=%s", dbS.Host, dbS.Port, dbS.Database, dbS.SSL, dbS.User, dbS.Pass)
}

//Функция создания нового воркера для работы с БД
//  @param
//     nameWorker    string               - имя воркера
//     config        cfg.Server           - массив структур конфигураций Server из файла конфигураций
//	   databases     []String             - имя database для подключения
//     logger        *logger.I_Logger_ext -
//     attemp        bool                 -
//     cntAttemp     uint                 -
//     timeAttemp    uint                 -
//  @return
// *   db_worker - указатель на объект воркера для работы с БД
//
func CreateDBWorker(nameWorker string, config *DbSettings, logger *gl.Logger, attemp bool, cntAttemp uint, timeAttemp uint) (*DbWorker, error) {
	if config == nil {
		return nil, errors.New("Nil config, Incorrect input param for NewDBWorker()")
	}
	newDBworker := &DbWorker{nameWorker, nil, config, logger, cntAttemp,
		timeAttemp, attemp}
	err := newDBworker.Connect()
	return newDBworker, err
}

// Создание пула коннекшенов к БД
func (pg *DbWorker) Connect() error {
	var err error
	dbW.Db, err = sqlx.Open("postgres", dbW.settings.String())
	if err != nil {
		(*dbW.logger).Error("connect()", "db_worker", "Can't open (get handle to database) to Db server!", dbW.settings.String(), err)
	}

	if err = dbW.Db.Ping(); err != nil {
		(*dbW.logger).Error("connect()", "db_worker", "Can't open connect (can't Ping) to Db server!", dbW.settings.String(), err)
	}
	return err
}

//Закрытие соединения
func (pg *PgDB) Close() {
	dbW.Db.Close()
	(*dbW.logger).Info("close()", "DbWorker",
		fmt.Sprintf("Connection to database %v:%v close()", dbW.settings.Host, dbW.settings.Database))
}

func (pg *PgDB) dbDefer(funcName string, fRecovery func(r interface{})) {
	if r := recover(); r != nil {
		(*dbW.logger).Error("Defer! For "+funcName, dbW.nameWorker, "PANIC!", r)
		fRecovery(r)
	}
}

// Функция обертка над execute. Запросы с ожиданием данных от БД. (SELECT и т.д. возращающие значения)
func (pg *PgDB) ExecuteQuery(nameFunc string, query string, args ...interface{}) (rows *sql.Rows, err error) {
	defer dbW.dbDefer(nameFunc+"-->"+"ExecuteQuery()", func(r interface{}) {
		sendEmail(fmt.Sprintf("Func: %v call this func: ExecuteQuery() \n Panic was! \n %v \n While do this SQL query: \n %v  \n With args: %v", nameFunc, r, query, args))
		err = errors.New(fmt.Sprintf("SQL query err: %v", r))
	})
	return dbW.execute(true, nameFunc, query, args...), err
}

// Функция обертка над execute. Запросы без ожидания данных от БД. (Update и т.д. не возращающие значения)
func (pg *PgDB) Execute(nameFunc string, query string, args ...interface{}) (err error) {
	defer dbW.dbDefer(nameFunc+"-->"+"Execute()", func(r interface{}) {
		sendEmail(fmt.Sprintf("Func: %v call this func: ExecuteQuery() \n Panic was! \n %v \n While do this SQL query: \n %v  \n With args: %v", nameFunc, r, query, args))
		err = errors.New(fmt.Sprintf("SQL query err: %v", r))
	})
	_ = dbW.execute(false, nameFunc, query, args...)
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
	if err = dbW.Db.Ping(); err != nil {
		(*dbW.logger).Error(nameFunc+"-->"+"execute()", "DbWorker", "Can't open connect (can't Ping) to Db server!", err)
		err = errors.New("no connect")
		panic(err)
	}
	for cnt := uint(0); cnt < dbW.cntAttemptRequest; cnt++ {
		(*dbW.logger).Debug(nameFunc+"-->"+"execute()", "DbWorker", fmt.Sprintf("Attemp execute SQL: %d", cnt))
		if returnValue { // TRUE == Execute_Query
			row, err = dbW.Db.Query(SQL, args...)
			if err != nil {
				(*dbW.logger).Log(gl.DB_FAIL, SQL, nameFunc+"-->"+"Execute_Query()", "Failed! Err:", err, "ARGS:", args)
			} else {
				(*dbW.logger).Log(gl.DB_OK, SQL, nameFunc+"-->"+"Execute_Query()", "SUCCESSES!", "ARGS:", args)
				return row
			}
			time.Sleep(time.Duration(dbW.timeAttemptRequest) * time.Second)
		} else { // FALSE == Execute
			var results sql.Result
			results, err = dbW.Db.Exec(SQL, args...)
			if err != nil {
				(*dbW.logger).Log(gl.DB_FAIL, SQL, nameFunc+"-->"+"Execute()", "Failed! Err:", err, "ARGS:", args)
			} else {
				(*dbW.logger).Log(gl.DB_OK, SQL, nameFunc+"-->"+"Execute()", "SUCCESSES!", "ARGS:", args)
				if rA, err := results.RowsAffected(); err == nil {
					(*dbW.logger).Info("Rows affected:", rA)
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

// Функция обертка над QueryX (sqlX). Запросы с ожиданием данных от БД.
func (pg *PgDB) ExecuteQueryX(nameFunc string, query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	defer pg.dbDefer(nameFunc+"-->"+"ExecuteQueryX()", func(r interface{}) {
		err = pg.Email.SendEmailByDefaultTemplate(fmt.Sprintf("Func: %v call this func: ExecuteQueryX() \n Panic was! \n %v \n While do this SQL query: \n %v  \n With args: %v", nameFunc, r, query, args))
		if err == nil {
			err = errors.New(fmt.Sprintf("SQL query err: %v", r))
		}
	})
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
	for cnt := uint(0); cnt < pg.CntAttemptRequest; cnt++ {
		(*pg.logger).Debug(nameFunc+"-->"+"executeX()", "DbWorker", fmt.Sprintf("Attemp execute SQL: %d", cnt))
		row, err = pg.db.Queryx(SQL, args...)
		if err != nil {
			(*pg.logger).Log(gl.DB_FAIL, SQL, nameFunc+"-->"+"ExecuteX QueryX()", "Failed! Err:", err, "ARGS:", args)
		} else {
			(*pg.logger).Log(gl.DB_OK, SQL, nameFunc+"-->"+"ExecuteX QueryX()", "SUCCESSES!", "ARGS:", args)
			return row
		}
		time.Sleep(time.Duration(pg.TimeAttemptRequest) * time.Second)
	}
	(*pg.logger).Error(nameFunc+"()-->"+"executeX()", "DbWorker", "All try estimates! Panic!", err)
	panic(err)
}

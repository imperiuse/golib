package transactionhelper

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Transaction is an interface that models the standard transaction in
// `database/sql`.
//
// To ensure `TxFn` funcs cannot commit or rollback a transaction (which is
// handled by `WithTransaction`), those methods are not included here.
type Transaction interface {
	sqlx.QueryerContext
	sqlx.PreparerContext
	sqlx.ExecerContext
}

// A Txfn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn func(Transaction) error

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func WithTransaction(db *sqlx.DB, fn TxFn) (err error) {
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			errR := tx.Rollback()
			err = errors.WithMessagef(err, "Panic in WithTransaction: %v. --> Rollback error: %v", p, errR)
		} else if err != nil {
			// something went wrong, rollback
			errR := tx.Rollback()
			err = errors.WithMessagef(err, "Err while execute fn. --> Rollback error: %v", errR)
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	err = fn(tx)

	return
}

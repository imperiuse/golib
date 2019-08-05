package transactionhelper

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// A Txfn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn func(*sqlx.Tx) error

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func WithTransaction(db *sqlx.DB, fn TxFn) (err error) {
	var tx *sqlx.Tx
	defer func() {
		if p := recover(); p != nil {
			err = errors.WithMessagef(err, "Panic was: %v", p)
		}
	}()
	func() {
		tx = db.MustBegin()

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
	}()
	return
}

// WithCtxTransaction creates a new transaction with ctx and handles rollback/commit based on the
// error object returned by the `TxFn`
func WithCtxTransaction(ctx context.Context, opt *sql.TxOptions, db *sqlx.DB, fn TxFn) (err error) {
	var tx *sqlx.Tx
	defer func() {
		if p := recover(); p != nil {
			err = errors.WithMessagef(err, "Panic was: %v", p)
		}
	}()
	func() {
		tx = db.MustBeginTx(ctx, opt)

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
	}()
	return
}

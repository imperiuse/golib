package sqlxhelper

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TxxI interface {
	// BeginTxx begins a transaction and returns an *sqlx.Tx instead of an *sql.Tx.
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// A Txfn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn = func(*sqlx.Tx) error

// WithTransaction execute [1...n] TxFn used one transaction
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back the transaction.
// Tx.Commit will return an error if the context is canceled.
// TxOptions holds the transaction options to be used in DB.BeginTx.
func WithTransaction(ctx context.Context, opt *sql.TxOptions, db TxxI, fn ...TxFn) error {
	tx, err := db.BeginTxx(ctx, opt)
	if err != nil {
		return err
	}

	// function used for panic control (defer inside)
	func() {
		defer func() {
			if p := recover(); p != nil {
				// a library panic occurred, rollback and repanic
				errR := tx.Rollback()
				err = errors.WithMessagef(err, "Panic in WithTransaction: %v. --> Rollback error: %v", p, errR)
			} else if err != nil {
				// something went wrong when, rollback
				errR := tx.Rollback()
				err = errors.WithMessagef(err, "Err while execute fn. --> Rollback error: %v", errR)
			} else {
				// all good, commit
				err = tx.Commit() // err!=nil when ctx is canceled
			}
		}()

		for _, f := range fn {
			err = f(tx)
			if err != nil {
				break // break loop, rollback in defer @see up
			}
		}

	}()

	return err
}

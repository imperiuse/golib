package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxxI interface which contain sqlx.BeginTxx func.
type TxxI interface {
	// BeginTxx begins a transaction and returns an *sqlx.Tx instead of an *sql.Tx.
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// TxFn is a function that will be called with an initialized `Transaction` object
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
		return fmt.Errorf("[WithTransaction] %w", err)
	}

	// function used for panic control (defer inside)
	func() {
		defer func() {
			if p := recover(); p != nil {
				if tx.Tx == nil {
					return
				}

				// a library panic occurred, rollback and repanic
				errR := tx.Rollback()
				err = fmt.Errorf("panic [WithTransaction]: %v. --> Rollback error: %v, %w", p, errR, err)

				return
			}

			// something went wrong when, rollback
			// err!=nil when ctx is canceled
			if err != nil {
				if tx.Tx == nil {
					return
				}

				errR := tx.Rollback()
				err = fmt.Errorf("err while Rollback. error: %v, %w", errR, err)

				return
			}

			// all good, commit
			err = tx.Commit()
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

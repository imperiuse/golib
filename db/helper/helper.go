package helper

import (
	"context"

	"github.com/imperiuse/golib/db/transaction"

	"github.com/jmoiron/sqlx"
)

// InsertAndGetLastID helper which Usefully for sql query like this
// query := `INSERT INTO table (col1, col2) VALUES ($1, $2) RETURNING id`.
func InsertAndGetLastID(ctx context.Context, lastInsertID any, query string, args ...any) transaction.TxFn {
	return func(t *sqlx.Tx) error {
		return t.QueryRowContext(ctx, query, args...).Scan(lastInsertID)
	}
}

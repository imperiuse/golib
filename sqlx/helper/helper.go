package helper

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// InsertAndGetLastID helper which Usefully for sql query like this
// query := `INSERT INTO table (col1, col2) VALUES ($1, $2) RETURNING id`.
func InsertAndGetLastID(ctx context.Context, lastInsertID interface{}, query string, args ...interface{}) TxFn {
	return func(t *sqlx.Tx) error {
		return t.QueryRowContext(ctx, query, args...).Scan(lastInsertID)
	}
}

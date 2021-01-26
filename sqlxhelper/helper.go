package sqlxhelper

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Usefully for sql query like this
// query := `INSERT INTO table (col1, col2) VALUES ($1, $2) RETURNING id`
func InsertAndGetLastID(ctx context.Context, lastInsertID *int64, query string, args ...interface{}) TxFn {
	return func(t *sqlx.Tx) error {
		return t.QueryRowContext(ctx, query, args...).Scan(lastInsertID)
	}
}

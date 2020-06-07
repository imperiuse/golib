package sqlxhelper

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Usefully for sql query like this
// query := `INSERT INTO table (col1, col2) VALUES ($1, $2) RETURNING id`
func InsertAndGetLastID(ctx context.Context, lastInsertID *int64, query string, args ...interface{}) TxFn {
	return func(t *sqlx.Tx) error {
		err := t.QueryRowContext(ctx, query, args...).Scan(lastInsertID)
		if err != nil {
			return fmt.Errorf("InsertAndGetLastID t.QueryRowContext err: %w  query: %s", err, query)
		}
		return nil
	}
}

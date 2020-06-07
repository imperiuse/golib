package sqlxhelper

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Usefully for sql query like this
// query := `INSERT INTO table (col1, col2) VALUES ($1, $2) RETURNING id`
func InsertAndGetLastID(lastInsertID *int64, query string, args ...interface{}) TxFn {
	return func(t *sqlx.Tx) error {
		stmt, err := t.Prepare(query)
		if err != nil {
			return fmt.Errorf("InsertAndGetLastID t.Prepare(query) err: %w  query: %s", err, query)
		}
		defer func() { _ = stmt.Close() }()

		err = stmt.QueryRow(args...).Scan(lastInsertID)
		if err != nil {
			return fmt.Errorf("InsertAndGetLastID stmt.QueryRow err: %w", err)
		}
		return nil
	}
}

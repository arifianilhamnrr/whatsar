package store

import (
	"database/sql"
	"fmt"
)

// SQLiteDSN returns a DSN tuned for concurrent access (WAL + busy timeout).
func SQLiteDSN(path string) string {
	return fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(60000)&_pragma=synchronous(NORMAL)",
		path,
	)
}

// ConfigureSQLite limits connections to 1 — required for modernc SQLite stability.
func ConfigureSQLite(conn *sql.DB) {
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
}
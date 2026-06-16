package store

import (
	"context"

	"modernc.org/sqlite"
)

func init() {
	sqlite.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, _ string) error {
		// Dipanggil tiap koneksi baru — pastikan pragma konsisten.
		pragmas := []string{
			"PRAGMA journal_mode=WAL",
			"PRAGMA foreign_keys=ON",
			"PRAGMA busy_timeout=60000",
			"PRAGMA synchronous=NORMAL",
			"PRAGMA temp_store=MEMORY",
		}
		ctx := context.Background()
		for _, p := range pragmas {
			if _, err := conn.ExecContext(ctx, p, nil); err != nil {
				return err
			}
		}
		return nil
	})
}
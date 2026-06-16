package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SessionRecord struct {
	ID        string
	Name      string
	WAJID     string
	Phone     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (db *DB) CreateSession(ctx context.Context, id, name string) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO sessions (id, name, status) VALUES (?, ?, 'created')`,
		id, name,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (db *DB) GetSession(ctx context.Context, id string) (*SessionRecord, error) {
	row := db.conn.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(wa_jid,''), COALESCE(phone,''), status, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	)
	return scanSession(row)
}

func (db *DB) ListSessions(ctx context.Context) ([]*SessionRecord, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, name, COALESCE(wa_jid,''), COALESCE(phone,''), status, created_at, updated_at
		 FROM sessions ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var out []*SessionRecord
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) UpdateSessionStatus(ctx context.Context, id, status string) error {
	_, err := db.conn.ExecContext(ctx,
		`UPDATE sessions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	return nil
}

func (db *DB) UpdateSessionPaired(ctx context.Context, id, waJID, phone, status string) error {
	_, err := db.conn.ExecContext(ctx,
		`UPDATE sessions SET wa_jid = ?, phone = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		waJID, phone, status, id,
	)
	if err != nil {
		return fmt.Errorf("update session paired: %w", err)
	}
	return nil
}

func (db *DB) DeleteSession(ctx context.Context, id string) error {
	_, err := db.conn.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func scanSession(row interface {
	Scan(dest ...any) error
}) (*SessionRecord, error) {
	var s SessionRecord
	err := row.Scan(&s.ID, &s.Name, &s.WAJID, &s.Phone, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}
	return &s, nil
}
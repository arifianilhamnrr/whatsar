package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type WebhookRecord struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id,omitempty"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"secret,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func (db *DB) CreateWebhook(ctx context.Context, wh *WebhookRecord) error {
	events, err := json.Marshal(wh.Events)
	if err != nil {
		return err
	}
	active := 0
	if wh.Active {
		active = 1
	}
	_, err = db.conn.ExecContext(ctx,
		`INSERT INTO webhooks (id, session_id, url, events, secret, active) VALUES (?, ?, ?, ?, ?, ?)`,
		wh.ID, nullIfEmpty(wh.SessionID), wh.URL, string(events), wh.Secret, active,
	)
	if err != nil {
		return fmt.Errorf("create webhook: %w", err)
	}
	return nil
}

func (db *DB) ListWebhooks(ctx context.Context, sessionID string) ([]*WebhookRecord, error) {
	var rows interface {
		Close() error
		Next() bool
		Scan(dest ...any) error
		Err() error
	}
	var err error

	if sessionID != "" {
		r, qerr := db.conn.QueryContext(ctx,
			`SELECT id, COALESCE(session_id,''), url, events, COALESCE(secret,''), active, created_at
			 FROM webhooks WHERE active = 1 AND (session_id = ? OR session_id IS NULL OR session_id = '')`,
			sessionID,
		)
		rows, err = r, qerr
	} else {
		r, qerr := db.conn.QueryContext(ctx,
			`SELECT id, COALESCE(session_id,''), url, events, COALESCE(secret,''), active, created_at
			 FROM webhooks WHERE active = 1`,
		)
		rows, err = r, qerr
	}
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	return scanWebhookRows(rows)
}

func scanWebhookRows(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*WebhookRecord, error) {
	var out []*WebhookRecord
	for rows.Next() {
		var wh WebhookRecord
		var eventsJSON string
		var active int
		if err := rows.Scan(&wh.ID, &wh.SessionID, &wh.URL, &eventsJSON, &wh.Secret, &active, &wh.CreatedAt); err != nil {
			return nil, err
		}
		wh.Active = active == 1
		_ = json.Unmarshal([]byte(eventsJSON), &wh.Events)
		out = append(out, &wh)
	}
	return out, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
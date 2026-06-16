package store

import (
	"context"
	"fmt"
	"time"
)

type MessageRecord struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Direction string    `json:"direction"`
	RemoteJID string    `json:"remote_jid"`
	Body      string    `json:"body"`
	WAMsgID   string    `json:"wa_msg_id,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (db *DB) SaveMessage(ctx context.Context, msg *MessageRecord) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO messages (id, session_id, direction, remote_jid, body, wa_msg_id, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.SessionID, msg.Direction, msg.RemoteJID, msg.Body, msg.WAMsgID, msg.Status,
	)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}

func (db *DB) ListMessages(ctx context.Context, sessionID string, limit, offset int) ([]*MessageRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, session_id, direction, remote_jid, COALESCE(body,''), COALESCE(wa_msg_id,''), status, created_at
		 FROM messages WHERE session_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var out []*MessageRecord
	for rows.Next() {
		var m MessageRecord
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Direction, &m.RemoteJID, &m.Body, &m.WAMsgID, &m.Status, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (db *DB) ListAllMessages(ctx context.Context, limit, offset int) ([]*MessageRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, session_id, direction, remote_jid, COALESCE(body,''), COALESCE(wa_msg_id,''), status, created_at
		 FROM messages ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list all messages: %w", err)
	}
	defer rows.Close()

	var out []*MessageRecord
	for rows.Next() {
		var m MessageRecord
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Direction, &m.RemoteJID, &m.Body, &m.WAMsgID, &m.Status, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}
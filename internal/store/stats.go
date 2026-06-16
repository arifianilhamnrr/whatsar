package store

import (
	"context"
	"database/sql"
	"fmt"
)

type MessageCounts struct {
	In  int
	Out int
}

func (db *DB) CountAllMessages(ctx context.Context) (int, error) {
	var n int
	err := db.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages`).Scan(&n)
	return n, err
}

func (db *DB) CountSessionMessages(ctx context.Context, sessionID string) (MessageCounts, error) {
	var c MessageCounts
	err := db.conn.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN direction = 'in' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN direction = 'out' THEN 1 ELSE 0 END), 0)
		 FROM messages WHERE session_id = ?`, sessionID,
	).Scan(&c.In, &c.Out)
	return c, err
}

func (db *DB) GetLastMessage(ctx context.Context, sessionID string) (*MessageRecord, error) {
	row := db.conn.QueryRowContext(ctx,
		`SELECT id, session_id, direction, remote_jid, COALESCE(body,''), COALESCE(wa_msg_id,''), status, created_at
		 FROM messages WHERE session_id = ? ORDER BY created_at DESC LIMIT 1`, sessionID,
	)
	var m MessageRecord
	err := row.Scan(&m.ID, &m.SessionID, &m.Direction, &m.RemoteJID, &m.Body, &m.WAMsgID, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("last message: %w", err)
	}
	return &m, nil
}
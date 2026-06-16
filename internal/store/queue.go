package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type QueueRecord struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	Recipient   string    `json:"recipient"`
	MsgType     string    `json:"msg_type"`
	Body        string    `json:"body,omitempty"`
	MediaURL    string    `json:"media_url,omitempty"`
	Caption     string    `json:"caption,omitempty"`
	ReplyTo     string    `json:"reply_to,omitempty"`
	QuotedText  string    `json:"quoted_text,omitempty"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	NextRetryAt time.Time `json:"next_retry_at"`
	Status      string    `json:"status"`
	LastError   string    `json:"last_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (db *DB) EnqueueMessage(ctx context.Context, rec *QueueRecord) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO message_queue
		 (id, session_id, recipient, msg_type, body, media_url, caption, reply_to, quoted_text,
		  attempts, max_attempts, next_retry_at, status, last_error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.SessionID, rec.Recipient, rec.MsgType, rec.Body, rec.MediaURL, rec.Caption,
		rec.ReplyTo, rec.QuotedText, rec.Attempts, rec.MaxAttempts, rec.NextRetryAt, rec.Status, rec.LastError,
	)
	if err != nil {
		return fmt.Errorf("enqueue message: %w", err)
	}
	return nil
}

func (db *DB) ListDueQueue(ctx context.Context, limit int) ([]*QueueRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, session_id, recipient, msg_type, COALESCE(body,''), COALESCE(media_url,''),
		        COALESCE(caption,''), COALESCE(reply_to,''), COALESCE(quoted_text,''),
		        attempts, max_attempts, next_retry_at, status, COALESCE(last_error,''), created_at
		 FROM message_queue
		 WHERE status = 'pending' AND next_retry_at <= CURRENT_TIMESTAMP
		 ORDER BY next_retry_at ASC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due queue: %w", err)
	}
	defer rows.Close()
	return scanQueueRows(rows)
}

func (db *DB) UpdateQueueAttempt(ctx context.Context, id string, attempts int, nextRetry time.Time, lastErr string) error {
	_, err := db.conn.ExecContext(ctx,
		`UPDATE message_queue SET attempts = ?, next_retry_at = ?, last_error = ?, status = 'pending'
		 WHERE id = ?`,
		attempts, nextRetry, lastErr, id,
	)
	return err
}

func (db *DB) MarkQueueDone(ctx context.Context, id string) error {
	_, err := db.conn.ExecContext(ctx,
		`UPDATE message_queue SET status = 'sent', last_error = '' WHERE id = ?`, id,
	)
	return err
}

func (db *DB) MarkQueueFailed(ctx context.Context, id, lastErr string) error {
	_, err := db.conn.ExecContext(ctx,
		`UPDATE message_queue SET status = 'failed', last_error = ? WHERE id = ?`, lastErr, id,
	)
	return err
}

func (db *DB) CountPendingQueue(ctx context.Context) (int, error) {
	var n int
	err := db.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM message_queue WHERE status = 'pending'`,
	).Scan(&n)
	return n, err
}

func scanQueueRows(rows *sql.Rows) ([]*QueueRecord, error) {
	var out []*QueueRecord
	for rows.Next() {
		var r QueueRecord
		if err := rows.Scan(
			&r.ID, &r.SessionID, &r.Recipient, &r.MsgType, &r.Body, &r.MediaURL, &r.Caption,
			&r.ReplyTo, &r.QuotedText, &r.Attempts, &r.MaxAttempts, &r.NextRetryAt, &r.Status, &r.LastError, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}
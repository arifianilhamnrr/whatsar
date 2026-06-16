package wa

import (
	"context"
	"math"
	"time"

	"github.com/whatsar/whatsar/internal/store"
)

func (m *Manager) StartQueueWorker(ctx context.Context) {
	go m.runQueueWorker(ctx)
}

func (m *Manager) runQueueWorker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.processQueueBatch(ctx)
		}
	}
}

func (m *Manager) processQueueBatch(ctx context.Context) {
	items, err := m.db.ListDueQueue(ctx, 15)
	if err != nil {
		m.log.Errorf("queue list: %v", err)
		return
	}
	for _, item := range items {
		m.processQueueItem(ctx, item)
	}
}

func (m *Manager) processQueueItem(ctx context.Context, item *store.QueueRecord) {
	out := OutgoingMessage{
		SessionID: item.SessionID,
		To:        item.Recipient,
		Type:      item.MsgType,
		Text:      item.Body,
		ImageURL:  item.MediaURL,
		Caption:   item.Caption,
		ReplyTo:   item.ReplyTo,
		QuotedText: item.QuotedText,
	}
	if item.MsgType == "image" && item.MediaURL == "" && item.Body != "" {
		out.ImageB64 = item.Body
		out.Text = ""
	}

	sess, err := m.Get(item.SessionID)
	if err != nil {
		m.failQueueItem(ctx, item, err.Error())
		return
	}

	waID, err := sess.sendOutgoing(ctx, out)
	if err == nil {
		_ = m.db.MarkQueueDone(ctx, item.ID)
		m.log.Infof("[queue] %s terkirim (%s)", item.ID[:8], waID)
		return
	}

	attempts := item.Attempts + 1
	if attempts >= item.MaxAttempts {
		_ = m.db.MarkQueueFailed(ctx, item.ID, err.Error())
		m.log.Errorf("[queue] %s gagal permanen: %v", item.ID[:8], err)
		return
	}

	delay := time.Duration(math.Pow(2, float64(attempts))) * 30 * time.Second
	next := time.Now().Add(delay)
	_ = m.db.UpdateQueueAttempt(ctx, item.ID, attempts, next, err.Error())
	m.log.Warnf("[queue] %s retry %d/%d dalam %s: %v", item.ID[:8], attempts, item.MaxAttempts, delay, err)
}

func (m *Manager) failQueueItem(ctx context.Context, item *store.QueueRecord, msg string) {
	attempts := item.Attempts + 1
	if attempts >= item.MaxAttempts {
		_ = m.db.MarkQueueFailed(ctx, item.ID, msg)
		return
	}
	delay := time.Duration(math.Pow(2, float64(attempts))) * 30 * time.Second
	_ = m.db.UpdateQueueAttempt(ctx, item.ID, attempts, time.Now().Add(delay), msg)
}
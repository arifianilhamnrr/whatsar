package wa

import (
	"context"
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"

	appstore "github.com/whatsar/whatsar/internal/store"
)

type QRInfo struct {
	Code        string `json:"code"`
	ImageBase64 string `json:"image_base64"`
	Status      string `json:"status"`
}

func (m *Manager) ListSessionInfos(ctx context.Context) ([]SessionInfo, error) {
	records, err := m.db.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]SessionInfo, 0, len(records))
	for _, rec := range records {
		out = append(out, m.buildSessionInfo(rec))
	}
	return out, nil
}

func (m *Manager) GetSessionInfo(ctx context.Context, id string) (*SessionInfo, error) {
	rec, err := m.db.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("session %s tidak ditemukan", id)
	}
	info := m.buildSessionInfo(rec)
	return &info, nil
}

func (m *Manager) GetQRInfo(ctx context.Context, id string) (*QRInfo, error) {
	sess, err := m.EnsureLoaded(ctx, id)
	if err != nil {
		return nil, err
	}

	code := sess.GetQR()
	st := sess.GetStatus()

	info := &QRInfo{
		Code:   code,
		Status: string(st),
	}

	if code != "" {
		png, err := qrcode.Encode(code, qrcode.Medium, 256)
		if err != nil {
			return nil, fmt.Errorf("generate qr: %w", err)
		}
		info.ImageBase64 = base64.StdEncoding.EncodeToString(png)
	}

	return info, nil
}

func (m *Manager) ListMessages(ctx context.Context, sessionID string, limit, offset int) ([]*appstore.MessageRecord, error) {
	return m.db.ListMessages(ctx, sessionID, limit, offset)
}

func (m *Manager) ListAllMessages(ctx context.Context, limit, offset int) ([]*appstore.MessageRecord, error) {
	return m.db.ListAllMessages(ctx, limit, offset)
}

func (m *Manager) RegisterWebhook(ctx context.Context, wh *appstore.WebhookRecord) error {
	return m.db.CreateWebhook(ctx, wh)
}

func (m *Manager) ListWebhooks(ctx context.Context, sessionID string) ([]*appstore.WebhookRecord, error) {
	return m.db.ListWebhooks(ctx, sessionID)
}
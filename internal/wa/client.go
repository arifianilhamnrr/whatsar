package wa

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	appstore "github.com/whatsar/whatsar/internal/store"
)

func newClient(device *store.Device, log waLog.Logger) *whatsmeow.Client {
	return whatsmeow.NewClient(device, log)
}

func (s *Session) SendText(ctx context.Context, to, text string) (string, error) {
	if !s.IsConnected() {
		return "", fmt.Errorf("session %s tidak terkoneksi", s.ID)
	}

	jid, err := parseRecipient(to)
	if err != nil {
		return "", err
	}

	resp, err := s.client.SendMessage(ctx, jid, &waProto.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	msgID := uuid.New().String()
	rec := &appstore.MessageRecord{
		ID:        msgID,
		SessionID: s.ID,
		Direction: "out",
		RemoteJID: jid.String(),
		Body:      text,
		WAMsgID:   resp.ID,
		Status:    "sent",
	}
	s.manager.dbAsync(func() {
		_ = s.manager.db.SaveMessage(context.Background(), rec)
	})

	return resp.ID, nil
}

func parseRecipient(to string) (types.JID, error) {
	to = strings.TrimSpace(to)
	if to == "" {
		return types.JID{}, fmt.Errorf("nomor tujuan kosong")
	}

	if strings.Contains(to, "@") {
		return types.ParseJID(to)
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.TrimPrefix(to, "+"))

	if digits == "" {
		return types.JID{}, fmt.Errorf("nomor tidak valid: %s", to)
	}

	return types.NewJID(digits, types.DefaultUserServer), nil
}
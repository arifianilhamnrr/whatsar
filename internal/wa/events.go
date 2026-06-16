package wa

import (
	"context"
	"log"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/whatsar/whatsar/internal/store"
)

func (s *Session) handleEvent(evt any) {
	switch v := evt.(type) {
	case *events.Connected:
		s.setStatus(StatusConnected)
		if s.client.Store.ID != nil {
			phone := s.client.Store.ID.User
			s.mu.Lock()
			s.phone = phone
			s.mu.Unlock()
			jid := s.client.Store.ID.String()
			s.manager.dbAsync(func() {
				_ = s.manager.db.UpdateSessionPaired(context.Background(), s.ID, jid, phone, string(StatusConnected))
			})
		}
		log.Printf("[%s] connected", s.ID)

	case *events.Disconnected:
		if s.status() != StatusStopped {
			s.setStatus(StatusReconnect)
			s.manager.dbAsync(func() {
				_ = s.manager.db.UpdateSessionStatus(context.Background(), s.ID, string(StatusReconnect))
			})
		}
		log.Printf("[%s] disconnected — auto-reconnect aktif", s.ID)

	case *events.LoggedOut:
		s.setStatus(StatusStopped)
		s.manager.dbAsync(func() {
			_ = s.manager.db.UpdateSessionStatus(context.Background(), s.ID, string(StatusStopped))
		})
		log.Printf("[%s] logged out", s.ID)

	case *events.Message:
		body := extractText(v)
		if body == "" {
			return
		}

		msgID := uuid.New().String()
		rec := &store.MessageRecord{
			ID:        msgID,
			SessionID: s.ID,
			Direction: "in",
			RemoteJID: v.Info.Sender.String(),
			Body:      body,
			WAMsgID:   v.Info.ID,
			Status:    "received",
		}
		s.manager.dbAsync(func() {
			_ = s.manager.db.SaveMessage(context.Background(), rec)
		})

		log.Printf("[%s] pesan masuk dari %s: %s", s.ID, v.Info.Sender.User, body)

		if s.manager.onMessage != nil {
			s.manager.onMessage(IncomingMessage{
				SessionID: s.ID,
				From:      v.Info.Sender.String(),
				Chat:      v.Info.Chat.String(),
				Body:      body,
				MessageID: v.Info.ID,
				IsGroup:   v.Info.IsGroup,
			})
		}

	case *events.Receipt:
		if v.Type == events.ReceiptTypeDelivered || v.Type == events.ReceiptTypeRead {
			log.Printf("[%s] receipt %s untuk %d pesan", s.ID, v.Type, len(v.MessageIDs))
		}
	}
}

func extractText(msg *events.Message) string {
	if msg.Message == nil {
		return ""
	}
	if c := msg.Message.GetConversation(); c != "" {
		return c
	}
	if ext := msg.Message.GetExtendedTextMessage(); ext != nil {
		return ext.GetText()
	}
	return ""
}
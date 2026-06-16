package wa

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Session struct {
	ID     string
	Name   string
	client *whatsmeow.Client

	mu     sync.RWMutex
	state  Status
	qrCode string
	phone  string

	manager *Manager
	log     waLog.Logger
}

func (s *Session) status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Session) setStatus(st Status) {
	s.mu.Lock()
	s.state = st
	s.mu.Unlock()
}

func (s *Session) GetStatus() Status {
	return s.status()
}

func (s *Session) GetQR() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.qrCode
}

func (s *Session) setQR(code string) {
	s.mu.Lock()
	s.qrCode = code
	s.state = StatusQRReady
	s.mu.Unlock()
	s.manager.dbAsync(func() {
		_ = s.manager.db.UpdateSessionStatus(context.Background(), s.ID, string(StatusQRReady))
	})
}

func (s *Session) IsConnected() bool {
	return s.client != nil && s.client.IsConnected() && s.client.IsLoggedIn()
}

func (s *Session) Phone() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phone != "" {
		return s.phone
	}
	if s.client != nil && s.client.Store.ID != nil {
		return s.client.Store.ID.User
	}
	return ""
}

func (s *Session) connect(ctx context.Context) error {
	if s.client.Store.ID == nil {
		return s.pairNew(ctx)
	}
	return s.connectExisting(ctx)
}

func (s *Session) pairNew(ctx context.Context) error {
	s.setStatus(StatusConnecting)
	s.manager.dbAsync(func() {
		_ = s.manager.db.UpdateSessionStatus(ctx, s.ID, string(StatusConnecting))
	})

	qrChan, err := s.client.GetQRChannel(ctx)
	if err != nil {
		s.setStatus(StatusFailed)
		return fmt.Errorf("qr channel: %w", err)
	}

	go s.watchQR(ctx, qrChan)

	if err := s.client.Connect(); err != nil {
		s.setStatus(StatusFailed)
		return fmt.Errorf("connect: %w", err)
	}

	return nil
}

func (s *Session) connectExisting(ctx context.Context) error {
	s.setStatus(StatusConnecting)
	s.manager.dbAsync(func() {
		_ = s.manager.db.UpdateSessionStatus(ctx, s.ID, string(StatusConnecting))
	})

	if err := s.client.Connect(); err != nil {
		s.setStatus(StatusFailed)
		return fmt.Errorf("connect: %w", err)
	}
	return nil
}

func (s *Session) watchQR(ctx context.Context, qrChan <-chan whatsmeow.QRChannelItem) {
	for item := range qrChan {
		switch item.Event {
		case whatsmeow.QRChannelEventCode:
			s.setQR(item.Code)
			s.log.Infof("QR updated (timeout %s)", item.Timeout)
		case "success":
			s.setStatus(StatusConnected)
			if s.client.Store.ID != nil {
				s.mu.Lock()
				s.phone = s.client.Store.ID.User
				s.mu.Unlock()
				jid := s.client.Store.ID.String()
				phone := s.client.Store.ID.User
				s.manager.dbAsync(func() {
					_ = s.manager.db.UpdateSessionPaired(ctx, s.ID, jid, phone, string(StatusConnected))
				})
			}
			s.log.Infof("pairing berhasil")
			return
		default:
			s.setStatus(StatusFailed)
			s.manager.dbAsync(func() {
				_ = s.manager.db.UpdateSessionStatus(ctx, s.ID, string(StatusFailed))
			})
			if item.Error != nil {
				s.log.Errorf("pairing gagal: %v", item.Error)
			} else {
				s.log.Errorf("pairing gagal: %s", item.Event)
			}
			return
		}
	}
}

func (s *Session) disconnect() {
	if s.client != nil {
		s.client.Disconnect()
	}
	s.setStatus(StatusStopped)
}

func (s *Session) gracefulDisconnect() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect()
	}
}
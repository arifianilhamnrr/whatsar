package wa

import "fmt"

type SlotStats struct {
	InMemory   int
	Paired     int
	PendingQR  int
	MaxPaired  int
	Unlimited  bool
}

func (m *Manager) SlotStats() SlotStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := SlotStats{MaxPaired: m.maxSess, Unlimited: m.maxSess <= 0}
	for _, s := range m.sessions {
		st := s.GetStatus()
		if st == StatusFailed || st == StatusStopped {
			continue
		}
		stats.InMemory++
		if s.isPaired() {
			stats.Paired++
		} else {
			stats.PendingQR++
		}
	}
	return stats
}

func (s *Session) isPaired() bool {
	return s.IsConnected() || (s.client != nil && s.client.Store.ID != nil)
}

func (m *Manager) reclaimMemorySlots() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, s := range m.sessions {
		st := s.GetStatus()
		if st == StatusFailed || st == StatusStopped {
			s.disconnect()
			delete(m.sessions, id)
		}
	}
}

func (m *Manager) checkCanAddSession() error {
	if m.maxSess <= 0 {
		return nil
	}

	m.reclaimMemorySlots()
	stats := m.SlotStats()

	if stats.Paired >= m.maxSess {
		return fmt.Errorf(
			"sudah ada %d nomor WhatsApp terhubung (maks %d) — hapus session lama di dashboard",
			stats.Paired, m.maxSess,
		)
	}

	maxMemory := m.maxSess + 1
	if stats.InMemory >= maxMemory {
		return fmt.Errorf(
			"slot penuh (%d/%d di memori) — hapus session yang menunggu QR di dashboard",
			stats.InMemory, maxMemory,
		)
	}

	return nil
}
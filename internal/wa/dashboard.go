package wa

import (
	"context"
	"time"

	appstore "github.com/whatsar/whatsar/internal/store"
)

type DashboardStats struct {
	TotalSessions    int  `json:"total_sessions"`
	Connected        int  `json:"connected"`
	Disconnected     int  `json:"disconnected"`
	WaitingQR        int  `json:"waiting_qr"`
	Failed           int  `json:"failed"`
	TotalMessages    int  `json:"total_messages"`
	MaxSessions      int  `json:"max_sessions"`
	SlotsUsed        int  `json:"slots_used"`
	PairedSlots      int  `json:"paired_slots"`
	PendingQRSlots   int  `json:"pending_qr_slots"`
	UnlimitedSlots   bool `json:"unlimited_slots"`
	InMemorySessions int  `json:"in_memory_sessions"`
}

type SessionInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone,omitempty"`
	WAJID         string    `json:"wa_jid,omitempty"`
	Status        string    `json:"status"`
	Connected     bool      `json:"connected"`
	InMemory      bool      `json:"in_memory"`
	MessagesIn    int       `json:"messages_in"`
	MessagesOut   int       `json:"messages_out"`
	LastMessage   string    `json:"last_message,omitempty"`
	LastMessageAt time.Time `json:"last_message_at,omitempty"`
	LastDirection string    `json:"last_direction,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DashboardData struct {
	Stats          DashboardStats `json:"stats"`
	Sessions       []SessionInfo  `json:"sessions"`
	RecentMessages []*appstore.MessageRecord `json:"recent_messages"`
}

func (m *Manager) GetDashboard(ctx context.Context) (*DashboardData, error) {
	records, err := m.db.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	totalMsgs, _ := m.db.CountAllMessages(ctx)

	slots := m.SlotStats()

	data := &DashboardData{
		Stats: DashboardStats{
			TotalSessions:    len(records),
			MaxSessions:      m.maxSess,
			SlotsUsed:        slots.InMemory,
			PairedSlots:      slots.Paired,
			PendingQRSlots:   slots.PendingQR,
			UnlimitedSlots:   slots.Unlimited,
			InMemorySessions: slots.InMemory,
			TotalMessages:    totalMsgs,
		},
		Sessions: make([]SessionInfo, 0, len(records)),
	}

	for _, rec := range records {
		info := m.buildSessionInfo(rec)

		switch info.Status {
		case string(StatusConnected):
			if info.Connected {
				data.Stats.Connected++
			} else {
				data.Stats.Disconnected++
			}
		case string(StatusQRReady), string(StatusConnecting), string(StatusCreated):
			data.Stats.WaitingQR++
		case string(StatusFailed), string(StatusStopped):
			data.Stats.Failed++
		default:
			if info.Connected {
				data.Stats.Connected++
			} else {
				data.Stats.Disconnected++
			}
		}

		data.Sessions = append(data.Sessions, info)
	}

	recent, _ := m.db.ListAllMessages(ctx, 8, 0)
	data.RecentMessages = recent

	return data, nil
}

func (m *Manager) buildSessionInfo(rec *appstore.SessionRecord) SessionInfo {
	info := SessionInfo{
		ID:        rec.ID,
		Name:      rec.Name,
		Phone:     rec.Phone,
		WAJID:     rec.WAJID,
		Status:    rec.Status,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}

	if counts, err := m.db.CountSessionMessages(context.Background(), rec.ID); err == nil {
		info.MessagesIn = counts.In
		info.MessagesOut = counts.Out
	}

	if last, err := m.db.GetLastMessage(context.Background(), rec.ID); err == nil && last != nil {
		info.LastMessage = last.Body
		info.LastMessageAt = last.CreatedAt
		info.LastDirection = last.Direction
	}

	if sess, err := m.Get(rec.ID); err == nil {
		info.InMemory = true
		info.Status = string(sess.GetStatus())
		info.Phone = sess.Phone()
		info.Connected = sess.IsConnected()
	}

	return info
}
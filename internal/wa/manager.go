package wa

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	appstore "github.com/whatsar/whatsar/internal/store"
)

type MessageHandler func(IncomingMessage)

type Manager struct {
	container *sqlstore.Container
	db        *appstore.DB
	sessions  map[string]*Session
	mu        sync.RWMutex
	log       waLog.Logger
	maxSess   int
	onMessage MessageHandler
	dataDir   string
	lockFile  *os.File
	dbQueue   chan func()
}

type Options struct {
	DataDir     string
	AppDBPath   string
	MaxSessions int
	LogLevel    string
	OnMessage   MessageHandler
}

func NewManager(opts Options) (*Manager, error) {
	if opts.DataDir == "" {
		opts.DataDir = "./data"
	}
	if opts.MaxSessions < 0 {
		opts.MaxSessions = 5
	}

	log := waLog.Stdout("WhatsApp", opts.LogLevel, true)

	lockFile, err := acquireInstanceLock(opts.DataDir)
	if err != nil {
		return nil, err
	}

	waDBPath := filepath.Join(opts.DataDir, "wa_store.db")
	waDB, err := sql.Open("sqlite", appstore.SQLiteDSN(waDBPath))
	if err != nil {
		releaseInstanceLock(lockFile, opts.DataDir)
		return nil, fmt.Errorf("open wa store: %w", err)
	}
	appstore.ConfigureSQLite(waDB)

	container := sqlstore.NewWithDB(waDB, "sqlite", log.Sub("Store"))
	if err := container.Upgrade(context.Background()); err != nil {
		waDB.Close()
		releaseInstanceLock(lockFile, opts.DataDir)
		return nil, fmt.Errorf("wa store upgrade: %w", err)
	}

	appDBPath := opts.AppDBPath
	if appDBPath == "" {
		appDBPath = filepath.Join(opts.DataDir, "whatsar.db")
	}
	db, err := appstore.Open(appDBPath)
	if err != nil {
		container.Close()
		releaseInstanceLock(lockFile, opts.DataDir)
		return nil, fmt.Errorf("app store: %w", err)
	}

	m := &Manager{
		container: container,
		db:        db,
		sessions:  make(map[string]*Session),
		log:       log,
		maxSess:   opts.MaxSessions,
		onMessage: opts.OnMessage,
		dataDir:   opts.DataDir,
		lockFile:  lockFile,
		dbQueue:   make(chan func(), 256),
	}
	go m.runDBQueue()

	if err := m.restoreSessions(context.Background()); err != nil {
		m.Close()
		return nil, err
	}

	return m, nil
}

func (m *Manager) restoreSessions(ctx context.Context) error {
	records, err := m.db.ListSessions(ctx)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.Status == string(StatusStopped) || rec.Status == string(StatusFailed) {
			continue
		}
		if rec.WAJID == "" {
			continue
		}
		if err := m.restoreSession(ctx, rec); err != nil {
			m.log.Warnf("gagal restore session %s: %v", rec.ID, err)
		}
	}
	return nil
}

func (m *Manager) restoreSession(ctx context.Context, rec *appstore.SessionRecord) error {
	jid, err := types.ParseJID(rec.WAJID)
	if err != nil {
		return fmt.Errorf("parse jid: %w", err)
	}

	device, err := m.container.GetDevice(ctx, jid)
	if err != nil {
		return err
	}
	if device == nil {
		return fmt.Errorf("device tidak ditemukan untuk %s", rec.WAJID)
	}

	_, err = m.startSession(ctx, rec.ID, rec.Name, device, rec.Phone)
	return err
}

func (m *Manager) Create(ctx context.Context, name string) (*Session, error) {
	if err := m.checkCanAddSession(); err != nil {
		return nil, err
	}

	if name == "" {
		name = "default"
	}

	id := uuid.New().String()
	device := m.container.NewDevice()

	if err := m.db.CreateSession(ctx, id, name); err != nil {
		return nil, err
	}

	sess, err := m.startSession(ctx, id, name, device, "")
	if err != nil {
		_ = m.db.DeleteSession(ctx, id)
		return nil, err
	}

	go func() {
		if err := sess.connect(context.Background()); err != nil {
			m.log.Errorf("[%s] connect error: %v", id, err)
		}
	}()

	return sess, nil
}

func (m *Manager) Connect(ctx context.Context, sessionID string) error {
	sess, err := m.Get(sessionID)
	if err != nil {
		return err
	}
	return sess.connect(ctx)
}

func (m *Manager) startSession(ctx context.Context, id, name string, device *store.Device, phone string) (*Session, error) {
	client := newClient(device, m.log.Sub(id[:8]))
	sess := &Session{
		ID:      id,
		Name:    name,
		client:  client,
		state:   StatusCreated,
		phone:   phone,
		manager: m,
		log:     m.log.Sub(id[:8]),
	}

	client.AddEventHandler(sess.handleEvent)

	m.mu.Lock()
	m.sessions[id] = sess
	m.mu.Unlock()

	if phone != "" {
		sess.setStatus(StatusConnected)
	}

	return sess, nil
}

func (m *Manager) AppDB() *appstore.DB {
	return m.db
}

func (m *Manager) Get(id string) (*Session, error) {
	m.mu.RLock()
	sess, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("session %s tidak ditemukan", id)
	}
	return sess, nil
}

// EnsureLoaded returns an in-memory session, loading from DB when needed.
func (m *Manager) EnsureLoaded(ctx context.Context, id string) (*Session, error) {
	if sess, err := m.Get(id); err == nil {
		return sess, nil
	}

	rec, err := m.db.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("session %s tidak ditemukan", id)
	}

	if rec.WAJID != "" {
		if err := m.restoreSession(ctx, rec); err != nil {
			return nil, err
		}
		return m.Get(id)
	}

	switch rec.Status {
	case string(StatusFailed), string(StatusStopped):
		return nil, fmt.Errorf("session sudah gagal — buat session baru")
	}

	if err := m.checkCanAddSession(); err != nil {
		return nil, err
	}

	device := m.container.NewDevice()
	sess, err := m.startSession(ctx, id, rec.Name, device, "")
	if err != nil {
		return nil, err
	}

	go func() {
		if err := sess.connect(context.Background()); err != nil {
			m.log.Errorf("[%s] connect error: %v", id, err)
		}
	}()

	return sess, nil
}

func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s)
	}
	return out
}

func (m *Manager) SendText(ctx context.Context, sessionID, to, text string) (string, error) {
	res, err := m.SendOutgoing(ctx, OutgoingMessage{
		SessionID: sessionID,
		To:        to,
		Text:      text,
		Type:      "text",
	})
	if err != nil {
		return "", err
	}
	return res.MessageID, nil
}

func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	if sess, err := m.Get(sessionID); err == nil {
		if sess.client.Store.ID != nil {
			_ = sess.client.Logout(ctx)
		} else {
			sess.disconnect()
		}
		m.mu.Lock()
		delete(m.sessions, sessionID)
		m.mu.Unlock()
	}

	return m.db.DeleteSession(ctx, sessionID)
}

func (m *Manager) runDBQueue() {
	for fn := range m.dbQueue {
		fn()
	}
}

func (m *Manager) dbAsync(fn func()) {
	m.dbQueue <- fn
}

func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.Unlock()

	for _, s := range sessions {
		s.gracefulDisconnect()
	}

	m.drainDBQueue(ctx)
}

func (m *Manager) drainDBQueue(ctx context.Context) {
	if m.dbQueue == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case fn, ok := <-m.dbQueue:
			if !ok {
				return
			}
			fn()
		default:
			return
		}
	}
}

func (m *Manager) Close() {
	if m.dbQueue != nil {
		close(m.dbQueue)
		m.dbQueue = nil
	}
	if m.container != nil {
		m.container.Close()
	}
	if m.db != nil {
		m.db.Close()
	}
	releaseInstanceLock(m.lockFile, m.dataDir)
}
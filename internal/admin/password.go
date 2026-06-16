package admin

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"

	appstore "github.com/whatsar/whatsar/internal/store"
)

const minPasswordLen = 6

var (
	ErrWrongPassword    = errors.New("password lama salah")
	ErrPasswordMismatch = errors.New("konfirmasi password tidak cocok")
	ErrPasswordTooShort = errors.New("password minimal 6 karakter")
)

type PasswordStore struct {
	db          *appstore.DB
	envPassword string
	mu          sync.RWMutex
	tokenCache  string
}

func NewPasswordStore(db *appstore.DB, envPassword string) *PasswordStore {
	return &PasswordStore{db: db, envPassword: envPassword}
}

func (p *PasswordStore) Verify(ctx context.Context, password string) bool {
	hash, err := p.db.GetSetting(ctx, appstore.SettingAdminPasswordHash)
	if err != nil || hash == "" {
		return subtle.ConstantTimeCompare([]byte(password), []byte(p.envPassword)) == 1
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (p *PasswordStore) ValidateSession(ctx context.Context, r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	expected, err := p.sessionToken(ctx)
	if err != nil || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(expected)) == 1
}

func (p *PasswordStore) Login(ctx context.Context, password string) (string, error) {
	if !p.Verify(ctx, password) {
		return "", ErrWrongPassword
	}
	if err := p.ensureHash(ctx, password); err != nil {
		return "", err
	}
	return p.rotateSession(ctx)
}

func (p *PasswordStore) ChangePassword(ctx context.Context, current, newPassword, confirm string) (string, error) {
	if !p.Verify(ctx, current) {
		return "", ErrWrongPassword
	}
	if newPassword != confirm {
		return "", ErrPasswordMismatch
	}
	if len(newPassword) < minPasswordLen {
		return "", ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	if err := p.db.SetSetting(ctx, appstore.SettingAdminPasswordHash, string(hash)); err != nil {
		return "", err
	}
	return p.rotateSession(ctx)
}

func (p *PasswordStore) sessionToken(ctx context.Context) (string, error) {
	p.mu.RLock()
	if p.tokenCache != "" {
		token := p.tokenCache
		p.mu.RUnlock()
		return token, nil
	}
	p.mu.RUnlock()

	token, err := p.db.GetSetting(ctx, appstore.SettingAdminSessionToken)
	if err != nil {
		return "", err
	}

	p.mu.Lock()
	p.tokenCache = token
	p.mu.Unlock()
	return token, nil
}

func (p *PasswordStore) ensureHash(ctx context.Context, password string) error {
	existing, err := p.db.GetSetting(ctx, appstore.SettingAdminPasswordHash)
	if err != nil {
		return err
	}
	if existing != "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return p.db.SetSetting(ctx, appstore.SettingAdminPasswordHash, string(hash))
}

func (p *PasswordStore) rotateSession(ctx context.Context) (string, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", err
	}
	if err := p.db.SetSetting(ctx, appstore.SettingAdminSessionToken, token); err != nil {
		return "", err
	}
	p.mu.Lock()
	p.tokenCache = token
	p.mu.Unlock()
	return token, nil
}

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
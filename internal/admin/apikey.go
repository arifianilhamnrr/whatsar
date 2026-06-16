package admin

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	appstore "github.com/whatsar/whatsar/internal/store"
)

const minAPIKeyLen = 16

var ErrAPIKeyTooShort = errors.New("API key minimal 16 karakter")

type APIKeyStore struct {
	db     *appstore.DB
	envKey string
	mu     sync.RWMutex
	cache  string
}

func NewAPIKeyStore(db *appstore.DB, envKey string) *APIKeyStore {
	return &APIKeyStore{db: db, envKey: envKey}
}

func (k *APIKeyStore) Get(ctx context.Context) (string, error) {
	k.mu.RLock()
	if k.cache != "" {
		key := k.cache
		k.mu.RUnlock()
		return key, nil
	}
	k.mu.RUnlock()

	stored, err := k.db.GetSetting(ctx, appstore.SettingAPIKey)
	if err != nil {
		return "", err
	}
	if stored == "" {
		stored = k.envKey
		if stored != "" {
			_ = k.db.SetSetting(ctx, appstore.SettingAPIKey, stored)
		}
	}

	k.mu.Lock()
	k.cache = stored
	k.mu.Unlock()
	return stored, nil
}

func (k *APIKeyStore) Validate(ctx context.Context, key string) bool {
	expected, err := k.Get(ctx)
	if err != nil || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(expected)) == 1
}

func (k *APIKeyStore) Regenerate(ctx context.Context) (string, error) {
	key, err := generateAPIKey()
	if err != nil {
		return "", err
	}
	return key, k.save(ctx, key)
}

func (k *APIKeyStore) Set(ctx context.Context, key string) error {
	key = trimSpace(key)
	if len(key) < minAPIKeyLen {
		return ErrAPIKeyTooShort
	}
	return k.save(ctx, key)
}

func (k *APIKeyStore) save(ctx context.Context, key string) error {
	if err := k.db.SetSetting(ctx, appstore.SettingAPIKey, key); err != nil {
		return err
	}
	k.mu.Lock()
	k.cache = key
	k.mu.Unlock()
	return nil
}

func (k *APIKeyStore) Mask(key string) string {
	if key == "" {
		return "—"
	}
	if len(key) <= 12 {
		return key[:2] + "••••"
	}
	return key[:8] + "••••" + key[len(key)-4:]
}

func generateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "wsk_" + hex.EncodeToString(b), nil
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
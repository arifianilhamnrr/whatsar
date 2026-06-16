//go:build tools

package tools

import (
	_ "github.com/go-chi/chi/v5"
	_ "github.com/google/uuid"
	_ "github.com/joho/godotenv"
	_ "go.mau.fi/whatsmeow"
	_ "golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)
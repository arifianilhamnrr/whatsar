package validate

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxTextLen       = 4096
	MaxCaptionLen    = 1024
	MaxFilenameLen   = 255
	MaxURLLen        = 2048
	MaxSessionName   = 64
	MaxWebhookSecret = 128
	DefaultListLimit = 50
	MaxListLimit     = 200
)

var (
	phoneDigits = regexp.MustCompile(`^\+?[0-9]{8,15}$`)
	jidPattern  = regexp.MustCompile(`^[0-9]+@(s\.whatsapp\.net|g\.us|lid)$`)
	filenamePat = regexp.MustCompile(`^[a-zA-Z0-9._ -]+$`)
)

func SessionID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("session_id wajib diisi")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("session_id harus UUID valid")
	}
	return nil
}

func SessionName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if len(name) > MaxSessionName {
		return fmt.Errorf("nama session maks %d karakter", MaxSessionName)
	}
	return nil
}

func Recipient(to string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("nomor tujuan wajib diisi")
	}
	if strings.Contains(to, "@") {
		if !jidPattern.MatchString(to) {
			return fmt.Errorf("format JID tidak valid")
		}
		return nil
	}
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.TrimPrefix(to, "+"))
	if !phoneDigits.MatchString(digits) {
		return fmt.Errorf("nomor harus 8-15 digit (format internasional 62xxx)")
	}
	return nil
}

func MessageType(raw string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(raw))
	if t == "" {
		return "text", nil
	}
	switch t {
	case "text", "image", "document":
		return t, nil
	default:
		return "", fmt.Errorf("type tidak didukung: %s (text, image, document)", raw)
	}
}

func Text(s string, required bool) error {
	s = strings.TrimSpace(s)
	if required && s == "" {
		return fmt.Errorf("text wajib diisi")
	}
	if len(s) > MaxTextLen {
		return fmt.Errorf("text maks %d karakter", MaxTextLen)
	}
	return nil
}

func Caption(s string) error {
	if len(s) > MaxCaptionLen {
		return fmt.Errorf("caption maks %d karakter", MaxCaptionLen)
	}
	return nil
}

func Filename(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) > MaxFilenameLen {
		return fmt.Errorf("filename maks %d karakter", MaxFilenameLen)
	}
	if !filenamePat.MatchString(s) {
		return fmt.Errorf("filename mengandung karakter tidak valid")
	}
	return nil
}

func HTTPURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url wajib diisi")
	}
	if len(raw) > MaxURLLen {
		return fmt.Errorf("url terlalu panjang")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("url tidak valid")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url harus http atau https")
	}
	return nil
}

func WebhookURL(raw string) error {
	return HTTPURL(raw)
}

func WebhookEvents(events []string) error {
	if len(events) == 0 {
		return nil
	}
	allowed := map[string]bool{"message.in": true}
	for _, e := range events {
		if !allowed[e] {
			return fmt.Errorf("event tidak didukung: %s", e)
		}
	}
	return nil
}

func WebhookSecret(s string) error {
	if len(s) > MaxWebhookSecret {
		return fmt.Errorf("secret maks %d karakter", MaxWebhookSecret)
	}
	return nil
}

func Pagination(limit, offset int) (int, int, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		return 0, 0, fmt.Errorf("limit maks %d", MaxListLimit)
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset tidak boleh negatif")
	}
	return limit, offset, nil
}

func Base64Payload(s string, field string) error {
	if s == "" {
		return nil
	}
	if len(s) > 90<<20 {
		return fmt.Errorf("%s terlalu besar", field)
	}
	return nil
}
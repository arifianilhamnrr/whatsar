package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/whatsar/whatsar/internal/wa"
)

const maxWebhookAttempts = 5

type Dispatcher struct {
	mgr    *wa.Manager
	client *http.Client
}

func NewDispatcher(mgr *wa.Manager) *Dispatcher {
	return &Dispatcher{
		mgr: mgr,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (d *Dispatcher) Handle(msg wa.IncomingMessage) {
	ctx := context.Background()
	hooks, err := d.mgr.ListWebhooks(ctx, msg.SessionID)
	if err != nil {
		log.Printf("[webhook] list error: %v", err)
		return
	}

	payload := map[string]any{
		"event":      "message.in",
		"session_id": msg.SessionID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"data": map[string]any{
			"from":       msg.From,
			"chat":       msg.Chat,
			"body":       msg.Body,
			"message_id": msg.MessageID,
			"type":       "text",
			"is_group":   msg.IsGroup,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, hook := range hooks {
		if !containsEvent(hook.Events, "message.in") {
			continue
		}
		go d.deliver(hook.URL, hook.Secret, body)
	}
}

func (d *Dispatcher) deliver(url, secret string, body []byte) {
	for attempt := 1; attempt <= maxWebhookAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Whatsar-Webhook/1.0")

		if secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			sig := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-Whatsar-Signature", fmt.Sprintf("sha256=%s", sig))
		}

		resp, err := d.client.Do(req)
		if err != nil {
			log.Printf("[webhook] %s attempt %d: %v", url, attempt, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				if attempt > 1 {
					log.Printf("[webhook] %s sukses attempt %d", url, attempt)
				}
				return
			}
			log.Printf("[webhook] %s attempt %d: status %d", url, attempt, resp.StatusCode)
		}

		if attempt < maxWebhookAttempts {
			delay := webhookBackoff(attempt)
			time.Sleep(delay)
		}
	}
	log.Printf("[webhook] %s gagal setelah %d attempt", url, maxWebhookAttempts)
}

func webhookBackoff(attempt int) time.Duration {
	sec := math.Pow(2, float64(attempt-1))
	return time.Duration(sec) * time.Second
}

func containsEvent(events []string, target string) bool {
	if len(events) == 0 {
		return target == "message.in"
	}
	for _, e := range events {
		if e == target || e == "*" {
			return true
		}
	}
	return false
}
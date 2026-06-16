package wa

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	appstore "github.com/whatsar/whatsar/internal/store"
)

type OutgoingMessage struct {
	SessionID  string
	To         string
	Type         string // text, image, document
	Text         string
	ImageURL     string
	ImageB64     string
	DocumentURL  string
	DocumentB64  string
	FileName     string
	MimeType     string
	Caption      string
	ReplyTo    string
	QuotedText string
	QueueRetry bool
}

type SendResult struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	Queued    bool   `json:"queued,omitempty"`
	QueueID   string `json:"queue_id,omitempty"`
}

func (m *Manager) SendOutgoing(ctx context.Context, msg OutgoingMessage) (*SendResult, error) {
	sess, err := m.Get(msg.SessionID)
	if err != nil {
		return nil, err
	}

	waID, err := sess.sendOutgoing(ctx, msg)
	if err == nil {
		return &SendResult{MessageID: waID, Status: "sent"}, nil
	}

	if !msg.QueueRetry {
		return nil, err
	}

	queueID, qerr := m.enqueueOutgoing(ctx, msg, err.Error())
	if qerr != nil {
		return nil, fmt.Errorf("%v (gagal enqueue: %v)", err, qerr)
	}
	return &SendResult{Status: "queued", Queued: true, QueueID: queueID}, nil
}

func (m *Manager) enqueueOutgoing(ctx context.Context, msg OutgoingMessage, lastErr string) (string, error) {
	msgType := msg.Type
	if msgType == "" {
		msgType = "text"
	}
	id := uuid.New().String()
	rec := &appstore.QueueRecord{
		ID:          id,
		SessionID:   msg.SessionID,
		Recipient:   msg.To,
		MsgType:     msgType,
		Body:        msg.Text,
		MediaURL:    msg.ImageURL,
		Caption:     msg.Caption,
		ReplyTo:     msg.ReplyTo,
		QuotedText:  msg.QuotedText,
		Attempts:    0,
		MaxAttempts: 5,
		NextRetryAt: time.Now().Add(30 * time.Second),
		Status:      "pending",
		LastError:   lastErr,
	}
	if msg.ImageB64 != "" {
		rec.Body = msg.ImageB64
	} else if msg.DocumentB64 != "" {
		rec.Body = msg.DocumentB64
	}
	if msg.DocumentURL != "" {
		rec.MediaURL = msg.DocumentURL
	}
	if msg.FileName != "" {
		if rec.Caption == "" {
			rec.Caption = msg.FileName
		}
	}
	if err := m.db.EnqueueMessage(ctx, rec); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Session) sendOutgoing(ctx context.Context, msg OutgoingMessage) (string, error) {
	if !s.IsConnected() {
		return "", fmt.Errorf("session %s tidak terkoneksi", s.ID)
	}

	jid, err := parseRecipient(msg.To)
	if err != nil {
		return "", err
	}

	msgType := strings.ToLower(strings.TrimSpace(msg.Type))
	if msgType == "" || msgType == "text" {
		return s.sendTextMessage(ctx, jid, msg)
	}
	if msgType == "image" {
		return s.sendImageMessage(ctx, jid, msg)
	}
	if msgType == "document" {
		return s.sendDocumentMessage(ctx, jid, msg)
	}
	return "", fmt.Errorf("tipe pesan tidak didukung: %s", msg.Type)
}

func (s *Session) sendTextMessage(ctx context.Context, jid types.JID, msg OutgoingMessage) (string, error) {
	waMsg := buildTextProto(msg.Text, msg.ReplyTo, msg.QuotedText)
	resp, err := s.client.SendMessage(ctx, jid, waMsg)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}
	s.saveOutgoing(jid, msg.Text, resp.ID)
	return resp.ID, nil
}

func (s *Session) sendImageMessage(ctx context.Context, jid types.JID, msg OutgoingMessage) (string, error) {
	data, err := loadImageBytes(msg.ImageURL, msg.ImageB64)
	if err != nil {
		return "", err
	}

	uploaded, err := s.client.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	caption := msg.Caption
	if caption == "" {
		caption = msg.Text
	}

	imageMsg := &waProto.ImageMessage{
		Caption:       proto.String(caption),
		Mimetype:      proto.String(http.DetectContentType(data)),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
	}
	if msg.ReplyTo != "" {
		imageMsg.ContextInfo = buildContextInfo(msg.ReplyTo, msg.QuotedText)
	}

	resp, err := s.client.SendMessage(ctx, jid, &waProto.Message{ImageMessage: imageMsg})
	if err != nil {
		return "", fmt.Errorf("send image: %w", err)
	}
	body := caption
	if body == "" {
		body = "[image]"
	}
	s.saveOutgoing(jid, body, resp.ID)
	return resp.ID, nil
}

func (s *Session) sendDocumentMessage(ctx context.Context, jid types.JID, msg OutgoingMessage) (string, error) {
	data, fileName, err := loadDocumentBytes(msg)
	if err != nil {
		return "", err
	}

	uploaded, err := s.client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return "", fmt.Errorf("upload document: %w", err)
	}

	mime := strings.TrimSpace(msg.MimeType)
	if mime == "" {
		mime = http.DetectContentType(data)
		if mime == "application/octet-stream" {
			mime = "application/pdf"
		}
	}

	docMsg := &waProto.DocumentMessage{
		FileName:      proto.String(fileName),
		Title:         proto.String(strings.TrimSuffix(fileName, filepath.Ext(fileName))),
		Caption:       proto.String(msg.Caption),
		Mimetype:      proto.String(mime),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
	}
	if msg.ReplyTo != "" {
		docMsg.ContextInfo = buildContextInfo(msg.ReplyTo, msg.QuotedText)
	}

	resp, err := s.client.SendMessage(ctx, jid, &waProto.Message{DocumentMessage: docMsg})
	if err != nil {
		return "", fmt.Errorf("send document: %w", err)
	}
	body := fileName
	if msg.Caption != "" {
		body = msg.Caption
	}
	s.saveOutgoing(jid, body, resp.ID)
	return resp.ID, nil
}

func buildTextProto(text, replyTo, quotedText string) *waProto.Message {
	if replyTo != "" {
		return &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: buildContextInfo(replyTo, quotedText),
			},
		}
	}
	return &waProto.Message{Conversation: proto.String(text)}
}

func buildContextInfo(replyTo, quotedText string) *waProto.ContextInfo {
	ci := &waProto.ContextInfo{StanzaID: proto.String(replyTo)}
	if quotedText != "" {
		ci.QuotedMessage = &waProto.Message{Conversation: proto.String(quotedText)}
	}
	return ci
}

func loadDocumentBytes(msg OutgoingMessage) ([]byte, string, error) {
	data, err := loadMediaBytes(msg.DocumentURL, msg.DocumentB64, 64<<20, "document_url atau document_base64 wajib untuk tipe document")
	if err != nil {
		return nil, "", err
	}
	fileName := strings.TrimSpace(msg.FileName)
	if fileName == "" {
		fileName = "document.pdf"
	}
	return data, fileName, nil
}

func loadImageBytes(url, b64 string) ([]byte, error) {
	data, err := loadMediaBytes(url, b64, 16<<20, "image_url atau image_base64 wajib untuk tipe image")
	return data, err
}

func loadMediaBytes(url, b64 string, maxSize int64, emptyErr string) ([]byte, error) {
	if b64 != "" {
		raw := b64
		if idx := strings.Index(raw, ","); idx >= 0 {
			raw = raw[idx+1:]
		}
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("decode image base64: %w", err)
		}
		return data, nil
	}
	if url == "" {
		return nil, fmt.Errorf("%s", emptyErr)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch image: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("image kosong")
	}
	return data, nil
}

func (s *Session) saveOutgoing(jid types.JID, body, waMsgID string) {
	rec := &appstore.MessageRecord{
		ID:        uuid.New().String(),
		SessionID: s.ID,
		Direction: "out",
		RemoteJID: jid.String(),
		Body:      body,
		WAMsgID:   waMsgID,
		Status:    "sent",
	}
	s.manager.dbAsync(func() {
		_ = s.manager.db.SaveMessage(context.Background(), rec)
	})
}

func parseRecipient(to string) (types.JID, error) {
	to = strings.TrimSpace(to)
	if to == "" {
		return types.JID{}, fmt.Errorf("nomor tujuan kosong")
	}

	if strings.Contains(to, "@") {
		return types.ParseJID(to)
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.TrimPrefix(to, "+"))

	if digits == "" {
		return types.JID{}, fmt.Errorf("nomor tidak valid: %s", to)
	}

	return types.NewJID(digits, types.DefaultUserServer), nil
}
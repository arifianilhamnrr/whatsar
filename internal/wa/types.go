package wa

type Status string

const (
	StatusCreated    Status = "created"
	StatusQRReady    Status = "qr_ready"
	StatusConnecting Status = "connecting"
	StatusConnected  Status = "connected"
	StatusReconnect  Status = "reconnecting"
	StatusDisconnect Status = "disconnected"
	StatusFailed     Status = "failed"
	StatusStopped    Status = "stopped"
)

type IncomingMessage struct {
	SessionID string
	From      string
	Chat      string
	Body      string
	MessageID string
	IsGroup   bool
}
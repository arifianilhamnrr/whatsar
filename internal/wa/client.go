package wa

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func newClient(device *store.Device, log waLog.Logger) *whatsmeow.Client {
	return whatsmeow.NewClient(device, log)
}
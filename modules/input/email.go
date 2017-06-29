package input

import (
	"github.com/ineiti/cybermind/broker"
)

/*
 * The email-module connects to an IMAP-server.
 */

type Email struct {
	Host string
	User string
	Pass string
}

func RegisterEmail(b *broker.Broker) error {
	return b.RegisterModule("email", NewEmail)
}

func NewEmail(b *broker.Broker, config []byte) broker.Module {
	return &Email{}
}

func (e *Email) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	return nil, nil
}

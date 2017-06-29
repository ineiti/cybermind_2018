package input

import (
	"github.com/ineiti/cybermind/broker"
)

/*
Handles SMS receiving and sending.
*/

type SMS struct {
	Path string
}

func RegisterSMS(b *broker.Broker) error {
	return b.RegisterModule("sms", NewSMS)
}

func NewSMS(b *broker.Broker, config []byte) broker.Module {
	return &SMS{}
}

func (s *SMS) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	return nil, nil
}
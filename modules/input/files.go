package input

import (
	"github.com/ineiti/cybermind/broker"
)

type Files struct {
	Path string
}

func RegisterFiles(b *broker.Broker) error {
	return b.RegisterModule("files", NewFiles)
}

func NewFiles(b *broker.Broker, config []byte) broker.Module {
	return &Files{}
}

func (f *Files) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	return nil, nil
}

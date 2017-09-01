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

func NewFiles(b *broker.Broker, id broker.ModuleID, msg *broker.Message) broker.Module {
	return &Files{}
}

func (f *Files) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	return nil, nil
}

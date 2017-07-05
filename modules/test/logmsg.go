package test

import "github.com/ineiti/cybermind/broker"

const ModuleLogger = "testLogger"

type Logger struct {
	Broker   *broker.Broker
	Messages []*broker.Message
}

func SpawnLogger(b *broker.Broker) *Logger {
	b.RegisterModule(ModuleLogger, NewLogger)
	b.SpawnModule(ModuleLogger, nil)
	return b.Modules[len(b.Modules)-1].(*Logger)
}

func NewLogger(b *broker.Broker, msg *broker.Message) broker.Module {
	return &Logger{
		Broker: b,
	}
}

func (l *Logger) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	l.Messages = append(l.Messages, m)
	return nil, nil
}

package broker

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1/log"
)

/*
The broker registers module-types and is able to spawn them as needed. It also
sends new messages to all registered modules.
*/
type ModuleRegistration func(b *Broker, cfg *Message) Module

var BrokerStop = "stop"

type Broker struct {
	ModuleTypes map[string]ModuleRegistration
	Modules     []Module
}

func NewBroker() *Broker {
	log.Lvl2("Starting broker")
	return &Broker{
		ModuleTypes: make(map[string]ModuleRegistration),
		Modules:     []Module{},
	}
}

func (b *Broker) RegisterModule(name string, reg ModuleRegistration) error {
	if _, ok := b.ModuleTypes[name]; ok {
		return errors.New("this name already exists")
	}
	log.Lvl3("Registered", name)
	b.ModuleTypes[name] = reg
	return nil
}

func (b *Broker) SpawnModule(module string, msg *Message) error {
	m, ok := b.ModuleTypes[module]
	if !ok {
		return errors.New("didn't find this module-type")
	}
	b.Modules = append(b.Modules, m(b, msg))
	log.Lvl2("Spawned module", module)
	log.Lvl3("Module-list is now", b.Modules)
	return nil
}

func (b *Broker) Start() error {
	if len(b.Modules) == 0 {
		return errors.New("cannot start without at least one active module")
	}
	log.Lvl2("Started broker, sending first nil-message")
	return b.BroadcastMessage(nil)
}

func (b *Broker) BroadcastMessage(msg *Message) error {
	log.Lvl3("asking all modules to process message", msg)
	var msgs []Message
	for name, m := range b.Modules {
		log.Lvl3("asking module", name, "to process message")
		ms, err := m.ProcessMessage(msg)
		if err != nil {
			return err
		}
		msgs = append(msgs, ms...)
	}
	for _, msg := range msgs {
		log.Lvl3("got new message", msg)
		if err := b.BroadcastMessage(&msg); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) NewObject(id ModuleID, data []byte) *Object {
	return &Object{
		ModuleID: id,
		GID:      random.Bytes(IDLen, random.Stream),
		Data:     data,
	}
}

func (b *Broker) Stop() error {
	log.Lvl2("Stopping broker")
	b.BroadcastMessage(&Message{
		Action: Action{Command: BrokerStop},
	})
	return nil
}

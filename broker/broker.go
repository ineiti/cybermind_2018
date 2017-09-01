package broker

import (
	"errors"

	"fmt"

	"gopkg.in/dedis/onet.v1/log"
)

/*
The broker registers module-types and is able to spawn them as needed. It also
sends new messages to all registered modules.
*/
type ModuleRegistration func(b *Broker, id ModuleID, cfg *Message) Module

var BrokerStop = "stop"

type Broker struct {
	ModuleTypes   map[string]ModuleRegistration
	ModuleEntries []ModuleEntry
}

type ModuleEntry struct {
	Module Module
	Name   string
	ID     ModuleID
}

func (me ModuleEntry) String() string {
	return fmt.Sprintf("%x:%s", me.ID[0:4], me.Name)
}

func NewBroker() *Broker {
	log.Lvl2("Starting broker")
	return &Broker{
		ModuleTypes: make(map[string]ModuleRegistration),
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

func (b *Broker) SpawnModule(module string, id ModuleID, msg *Message) (Module, error) {
	m, ok := b.ModuleTypes[module]
	if !ok {
		return nil, errors.New("didn't find this module-type")
	}
	if id == nil {
		id = NewModuleID()
	}
	mod := m(b, id, msg)
	if mod == nil {
		return nil, errors.New("Couldn't spawn module: " + module)
	}
	me := ModuleEntry{
		Module: mod,
		Name:   module,
		ID:     id,
	}
	b.ModuleEntries = append(b.ModuleEntries, me)
	log.Lvl2("Spawned module", me)
	log.Lvl3("Module-list is now", b.ModuleEntries)
	return mod, nil
}

func (b *Broker) Start() error {
	if len(b.ModuleEntries) == 0 {
		return errors.New("cannot start without at least one active module")
	}
	log.Lvl2("Started broker, sending first nil-message")
	return b.BroadcastMessage(nil)
}

func (b *Broker) BroadcastMessage(msg *Message) error {
	if msg != nil && msg.ID == nil {
		log.Error("empty id for", log.Stack())
	}
	log.Lvl3("asking all modules", b.ModuleEntries, "to process message", msg)
	var msgs []Message
	for _, m := range b.ModuleEntries {
		log.Lvl3("asking module", m, "to process message", msg)
		ms, err := m.Module.ProcessMessage(msg)
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

func (b *Broker) BroadcastMessages(msgs []Message) error {
	for _, m := range msgs {
		if err := b.BroadcastMessage(&m); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) Stop() error {
	log.Lvl2("Stopping broker")
	b.BroadcastMessage(&Message{
		ID:     NewMessageID(),
		Action: Action{Command: BrokerStop},
	})
	return nil
}

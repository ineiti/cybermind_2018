package base

import (
	"path/filepath"

	"bytes"

	"errors"

	"github.com/ineiti/cybermind/broker"
	"gopkg.in/dedis/onet.v1/log"
)

const ModuleConfig = "config"

var ModuleIDConfig = []byte{0, 0, 0, 0}

var ConfigSpawn = broker.SubDomain("spawn", "config")
var ConfigData = broker.SubDomain("data", "config")
var ConfigModule = broker.SubDomain("module", "config")

type Config struct {
	Broker *broker.Broker
}

func RegisterConfig(b *broker.Broker) error {
	err := b.RegisterModule(ModuleConfig, NewConfig)
	if err != nil {
		return err
	}
	return b.SpawnModule(ModuleConfig, nil)
}

func NewConfig(b *broker.Broker, msg *broker.Message) broker.Module {
	return &Config{
		Broker: b,
	}
}

func (c *Config) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	if m == nil {
		log.Lvl2("Reading config and spawning modules")
		if err := c.SpawnConfigs(); err != nil {
			return nil, err
		}
	} else {
		log.LLvl3("Got message", m)
		if m.Action.Command == ConfigSpawn {
			err := c.ProcessSpawn(m)
			if err != nil {
				return nil, err
			}
		} else if len(m.Objects) == 1 &&
			bytes.Compare(m.Objects[0].ModuleID, ModuleIDConfig) == 0 {
			module := m.Tags.GetLastValue(ConfigModule)
			if module == nil {
				return nil, errors.New("no module-tag for configuration-object")
			}
			return nil, c.Broker.SpawnModule(module.Value, m)
		}
	}
	return nil, nil
}

func (c *Config) ProcessSpawn(msg *broker.Message) error {
	obj := c.Broker.NewObject(ModuleIDConfig, broker.NewModuleID())
	tags := StorageSaveTags.
		Add(broker.NewTag(ConfigData, msg.Action.Arguments["config"])).
		Add(broker.NewTag(ConfigModule, msg.Action.Arguments["module"]))
	newMsg := &broker.Message{
		Objects: []broker.Object{*obj},
		Tags:    tags,
	}
	return c.Broker.BroadcastMessage(newMsg)
}

func (c *Config) SpawnConfigs() error {
	return c.Broker.BroadcastMessage(&broker.Message{
		Action: broker.Action{
			Command: StorageSearchObject,
			Arguments: map[string]string{
				"module": string(ModuleIDConfig),
				"tags":   "*",
			},
		},
	})
	return nil
}

func configFile() string {
	return filepath.Join(broker.ConfigPath, "cybermind.toml")
}

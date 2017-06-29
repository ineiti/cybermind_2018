package data

import (
	"path/filepath"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/ineiti/cybermind/broker"
	"gopkg.in/dedis/onet.v1/log"
)

const moduleConfig = "config"

type ConfigModule struct {
	Name   string
	Config []byte
}

type ConfigModules struct {
	Modules []ConfigModule
}

type Config struct {
	Broker        *broker.Broker
	ConfigModules *ConfigModules
}

func RegisterConfig(b *broker.Broker) error {
	err := b.RegisterModule(moduleConfig, NewConfig)
	if err != nil {
		return err
	}
	return b.SpawnModule(moduleConfig, nil)
}

func NewConfig(b *broker.Broker, config []byte) broker.Module {
	return &Config{
		Broker:        b,
		ConfigModules: &ConfigModules{},
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
		if len(m.Tags) > 0 {
			err := c.ProcessActions(m.Actions)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (c *Config) ProcessActions(actions []broker.Action) error {
	for _, a := range actions {
		switch a.Command {
		case broker.SubDomain("spawn", "broker"):
			log.Lvl2("Storing config", a.Arguments["name"])
			cm := ConfigModule{
				Name:   a.Arguments["name"],
				Config: []byte(a.Arguments["config"]),
			}
			c.ConfigModules.Modules = append(c.ConfigModules.Modules, cm)
			if err := c.Save(); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func (c *Config) Save() error {
	f, err := os.OpenFile(configFile(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	log.LLvl3("Saving configuration", c.ConfigModules, configFile())
	return toml.NewEncoder(f).Encode(c.ConfigModules)
}

func (c *Config) SpawnConfigs() error {
	confMod := &ConfigModules{}
	_, err := toml.DecodeFile(configFile(), confMod)
	if err != nil {
		log.Warn(err)
		return nil
	}
	log.LLvl3("Loaded configuration", confMod, configFile())
	for _, conf := range confMod.Modules {
		log.Lvl2("Spawing module", conf.Name)
		c.Broker.SpawnModule(conf.Name, conf.Config)
	}
	return nil
}

func configFile() string {
	return filepath.Join(broker.ConfigPath, "cybermind.toml")
}

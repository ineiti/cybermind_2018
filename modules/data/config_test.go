package data

import (
	"testing"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/dedis/onet.v1/log"
)

func TestRegisterToml(t *testing.T) {
	tt := initTest(0)
	require.NotNil(t, tt.Toml)
}

func TestToml_ProcessMessage(t *testing.T) {
	tt := initTest(0)
	log.ErrFatal(tt.Broker.Start())
	log.ErrFatal(tt.Broker.SpawnModule(test.ModuleTestInput, nil))
	log.ErrFatal(tt.Broker.Stop())
	cf := configFile

	tt = initTest(0)
	configFile = cf
	log.ErrFatal(tt.Broker.Start())
	require.Equal(t, 2, len(tt.Broker.Modules))
}

type testToml struct {
	Broker *broker.Broker
	Toml   *Config
}

func initTest(cmd int) *testToml {
	tt := &testToml{
		Broker: broker.NewBroker(),
	}
	log.ErrFatal(RegisterConfig(tt.Broker))
	tt.Toml = tt.Broker.Modules[0].(*Config)
	log.ErrFatal(test.RegisterTestInput(tt.Broker))
	return tt
}

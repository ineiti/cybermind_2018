package input

import (
	"testing"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/base"
	"github.com/ineiti/cybermind/modules/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/dedis/onet.v1/log"
)

func TestEmail_Start(t *testing.T) {
	te := initTest(0)
	te.Broker.Start()
	defer te.Broker.Stop()
	log.ErrFatal(base.SpawnModule(te.Broker, NameEmail,
		"Plain://username:password@localhost"))

	require.Equal(t, 4, len(te.Broker.Modules))
	log.Lvl2(te.Logger.Messages)
}

func TestEmail_Restart(t *testing.T) {
	te := initTest(1)
	log.Lvl2(te.Broker.Modules)
	id := te.Broker.Modules[len(te.Broker.Modules)-1].(*Email).moduleid
	te.Broker.Stop()

	log.Lvl2(te.Logger.Messages)
	te = initTest(0)
	te.Broker.Start()
	defer te.Broker.Stop()
	log.Lvlf2("%+v", te.Broker.Modules)
	log.Lvl2(te.Logger.Messages)
	require.Equal(t, 4, len(te.Broker.Modules))
	require.Equal(t, id, te.Broker.Modules[len(te.Broker.Modules)-1].(*Email).moduleid)
}

type testEmail struct {
	Broker *broker.Broker
	Logger *test.Logger
}

func initTest(cmd int) *testEmail {
	te := &testEmail{
		Broker: broker.NewBroker(),
	}
	te.Logger = test.SpawnLogger(te.Broker)
	log.ErrFatal(base.RegisterStorage(te.Broker))
	log.ErrFatal(base.RegisterConfig(te.Broker))
	log.ErrFatal(test.RegisterTestInput(te.Broker))
	log.ErrFatal(RegisterEmail(te.Broker))

	for i := 1; i <= cmd; i++ {
		switch i {
		case 1:
			te.Broker.Start()
			log.ErrFatal(base.SpawnModule(te.Broker, NameEmail,
				"Plain://username:password@localhost"))
		}
	}
	return te
}

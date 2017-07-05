package base

import (
	"testing"

	"github.com/ineiti/cybermind/broker"
	"gopkg.in/dedis/onet.v1/log"
)

func TestMain(m *testing.M) {
	broker.TempConfigPath()
	log.MainTest(m)
}

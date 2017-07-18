package input

import (
	"regexp"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/base"
	"gopkg.in/dedis/onet.v1/log"
)

/*
 * The email-module connects to an IMAP-server.
 */

var configRegexp = regexp.MustCompile("(.*)://(.*[^//]):(.*[^//])@(.*)")

var NameEmail = "email"

type Email struct {
	Host           string
	User           string
	Pass           string
	ConnectionType ConnectionType
	moduleid       broker.ModuleID
}

type ConnectionType int

const (
	CTPlain ConnectionType = iota
	CTTLS
	CTTLSIgnore
)

func RegisterEmail(b *broker.Broker) error {
	return b.RegisterModule(NameEmail, NewEmail)
}

func NewEmail(b *broker.Broker, msg *broker.Message) broker.Module {
	conf := msg.Tags.GetLastValue(base.ConfigData).Value
	configs := configRegexp.FindStringSubmatch(conf)
	log.Printf("Configs is: %v - %+v", msg.Tags, configs)
	if len(configs) != 5 {
		return nil
	}
	email := &Email{
		Host:     configs[4],
		User:     configs[2],
		Pass:     configs[3],
		moduleid: []byte(msg.Tags.GetLastValue("module_id").Value),
	}
	switch configs[1] {
	case "Plain":
		email.ConnectionType = CTPlain
	case "TLS":
		email.ConnectionType = CTTLS
	case "TLSIgnore":
		email.ConnectionType = CTTLSIgnore
	default:
		log.Error("wrong connection type: " + configs[0])
		return nil
	}
	return email
}

func (e *Email) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	if m == nil {

	}
	return nil, nil
}

func (e *Email) GetNew() {}

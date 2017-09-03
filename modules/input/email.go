package input

import (
	"regexp"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/base"
	"gopkg.in/dedis/onet.v1/log"
)

/*
 * The email-module connects to an IMAP-server.
 */

const ModuleEmail = "email"

type Email struct {
	Host           string
	User           string
	Pass           string
	ConnectionType ConnectionType
	Mails          [][]byte
	moduleid       broker.ModuleID
	searchid       broker.MessageID
	client         *client.Client
}

type ConnectionType int

const (
	CTPlain ConnectionType = iota
	CTTLS
	CTTLSIgnore
)

func RegisterEmail(b *broker.Broker) error {
	return b.RegisterModule(ModuleEmail, NewEmail)
}

var configRegexp = regexp.MustCompile("(.*)://(.*[^//]):(.*[^//])@(.*)")

func NewEmail(b *broker.Broker, id broker.ModuleID, msg *broker.Message) broker.Module {
	conf := msg.Tags.GetLastValue(base.ConfigData).Value
	configs := configRegexp.FindStringSubmatch(conf)
	log.Lvlf1("Configs is: %v - %+v", msg.Tags, configs)
	if len(configs) != 5 {
		return nil
	}
	email := &Email{
		Host:     configs[4],
		User:     configs[2],
		Pass:     configs[3],
		moduleid: id,
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
		search := []broker.Message{
			base.StorageSearchObject(broker.NewKeyValue("module_id", string(e.moduleid))),
		}
		e.searchid = search[0].ID
		var err error
		e.client, err = client.Dial(e.Host)
		if err != nil {
			return nil, err
		}
		log.Lvl2("Connected to", e.Host)

		// Login
		if err := e.client.Login(e.User, e.Pass); err != nil {
			return nil, err
		}
		log.Lvl2("Logged in", e.User)
		return search, nil
	} else if m.Action.Command == base.StorageActionSearchResult {
		if m.Action.Arguments["search_id"] == string(e.searchid) {
			log.Lvl2("Got a search result:", m.Objects)
			e.AddMails(m.Objects)
			e.GetNew()
		}
	} else if m.Action.Command == broker.BrokerStop {
		if e.client != nil {
			e.client.Close()
		}
	}
	return nil, nil
}

func (e *Email) AddMails(objs []broker.Object) {
	for _, o := range objs {
		e.Mails = append(e.Mails, o.Data)
	}
}

func (e *Email) GetNew() (*broker.Message, error) {
	// Select INBOX
	mbox, err := e.client.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Lvl1("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- e.client.Fetch(seqset, []string{imap.EnvelopeMsgAttr}, messages)
	}()

	for msg := range messages {
		log.Lvl1("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		return nil, err
	}
	return nil, nil
}

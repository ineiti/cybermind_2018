package input

import (
	"regexp"

	"crypto/sha256"

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
	Mails          map[string]*imap.Message
	moduleid       broker.ModuleID
	searchid       broker.MessageID
	client         *client.Client
	lastSeq        uint32
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

var configRegexp = regexp.MustCompile("(.*)://(.*[^//]):(.*[^//])#(.*)")

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
		Mails:    map[string]*imap.Message{},
		moduleid: id,
		lastSeq:  1,
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
	}
	switch m.Action.Command {
	case base.StorageActionSearchResult:
		if m.Action.Arguments["search_id"] == string(e.searchid) {
			log.Lvl2("Got a search result:", m.Objects)
			e.AddMails(m.Objects)
			msg, err := e.GetNew()
			if err != nil {
				return nil, err
			}
			e.AddMails(msg.Objects)
		}
	case broker.BrokerStop:
		if e.client != nil {
			e.client.Close()
		}
	case InputActionPoll:
		if e.client != nil {
			msg, err := e.GetNew()
			if err == nil {
				e.AddMails(msg.Objects)
			}
			return []broker.Message{*msg}, err
		}
	}
	return nil, nil
}

func (e *Email) AddMails(objs []broker.Object) {
	for _, o := range objs {
		e.Mails[hashString(o.Data)] = BytesToEmail(o.Data)
	}
}

func (e *Email) GetNew() (msg *broker.Message, err error) {
	// Select INBOX
	mbox, err := e.client.Select("INBOX", false)
	if err != nil {
		return nil, err
	}
	log.LLvl3("Flags for INBOX:", mbox.Flags, e.lastSeq)

	from := uint32(1)
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- e.client.Fetch(seqset, []string{imap.EnvelopeMsgAttr,
			imap.UidMsgAttr, imap.BodyMsgAttr, imap.BodyStructureMsgAttr, imap.FlagsMsgAttr}, messages)
	}()

	msg = broker.NewMessage()
	for m := range messages {
		b := EmailToBytes(m)
		if _, exists := e.Mails[hashString(b)]; !exists {
			log.LLvl3("*New*", m.SeqNum, m.Envelope.Subject)
			msg.Objects = append(msg.Objects,
				*broker.NewObject(e.moduleid, b))
		} else {
			log.LLvl3("-old-", m.SeqNum, m.Envelope.Subject)
		}
		log.Printf("%#v", m.BodyStructure)
		e.lastSeq = m.SeqNum + 1
	}

	if e.lastSeq == uint32(28) {
		log.Print("Deleting message")
		ss := new(imap.SeqSet)
		ss.AddNum(uint32(2))
		log.ErrFatal(e.client.Store(ss, imap.RemoveFlags, []interface{}{"\\Deleted"}, nil))
	}

	err = <-done
	return
}

func EmailToBytes(m *imap.Message) []byte {
	return []byte(m.Envelope.Subject)
	return nil
}

func BytesToEmail(b []byte) *imap.Message {
	m := imap.NewMessage(0, nil)
	m.Envelope = &imap.Envelope{
		Subject: string(b),
	}
	return m
}

func hashString(b []byte) string {
	h := sha256.Sum256(b)
	return string(h[:])
}

package input

import (
	"testing"

	"bytes"

	"time"

	"strings"

	"fmt"

	"math/rand"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/server"
	"github.com/ineiti/go-imap/backend"
	"gopkg.in/dedis/onet.v1/log"
)

func TestGoImap(t *testing.T) {
	serv := startIMAPServer()
	log.Lvl1("Connecting to server...")

	// Connect to server
	c, err := client.Dial("localhost:1143")
	if err != nil {
		log.Fatal(err)
	}
	log.Lvl1("Connected")

	// Login
	if err := c.Login("username", "password"); err != nil {
		log.Fatal(err)
	}
	log.Lvl1("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Lvl1("Mailboxes:")
	for m := range mailboxes {
		log.Lvl1("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Lvl1("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []string{imap.EnvelopeMsgAttr}, messages)
	}()

	log.Lvl1("Last 4 messages:")
	for msg := range messages {
		log.Lvl1("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
	c.Logout()
	<-c.LoggedOut()
	c.Close()
	serv.Close()
	log.Lvl1("Done!")
}

func startIMAPServer() *server.Server {
	log.Lvl3("Starting server")
	// Create a memory backend
	be := memory.New()

	// Create a new server
	s := server.New(be)
	s.Addr = ":1143"
	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Lvl1("Starting IMAP server at localhost:1143")
	go func() {
		if err := s.ListenAndServe(); err != nil {
			if strings.Contains(err.Error(), "use of closed") {
				log.Lvl2("Quitting server")
				return
			}
			log.Fatal(err)
		}
		log.Print()
	}()

	addSpam(s, "2nd message", "Hi there :)")
	return s
}

func addSpam(s *server.Server, subject, body string) {
	id := fmt.Sprintf("%09d@localhost", rand.Int31())
	addMsg(s, "contact@example.org", "contact@example.org", subject,
		"Wed, 11 May 2017 14:31:59 +0000", id, "text/plain",
		body)
}

func addMsg(s *server.Server, from, to, subject, date, id, ct, body string) {
	mb := getInbox(s)
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\nDate: %s\n"+
		"Message-ID: <%s/>\nContent-Type: %s\n\n%s",
		from, to, subject, date, id, ct, body)
	err := mb.CreateMessage([]string{"\\Seen"}, time.Now(), imap.Literal(bytes.NewBufferString(msg)))
	if err != nil {
		log.Fatal(err)
	}
}

func getInbox(s *server.Server) backend.Mailbox {
	user, err := s.Backend.Login("username", "password")
	if err != nil {
		log.Fatal(err)
	}
	mb, err := user.GetMailbox("INBOX")
	if err != nil {
		log.Fatal(err)
	}
	return mb
}

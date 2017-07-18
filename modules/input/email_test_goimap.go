package input

import (
	"testing"

	"bytes"

	"time"

	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/server"
	"gopkg.in/dedis/onet.v1/log"
)

func TestGoImap(t *testing.T) {
	serv := startServer()
	log.Print("Connecting to server...")

	// Connect to server
	c, err := client.Dial("localhost:1143")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("username", "password"); err != nil {
		log.Fatal(err)
	}
	log.Print("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Print("Mailboxes:")
	for m := range mailboxes {
		log.Print("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Flags for INBOX:", mbox.Flags)

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

	log.Print("Last 4 messages:")
	for msg := range messages {
		log.Print("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	serv.Close()
	log.Print("Done!")
}

func startServer() *server.Server {
	log.Print("Starting server")
	// Create a memory backend
	be := memory.New()
	user, err := be.Login("username", "password")
	if err != nil {
		log.Fatal(err)
	}
	mb, err := user.GetMailbox("INBOX")
	if err != nil {
		log.Fatal(err)
	}
	msg := `From: contact@example.org
To: contact@example.org
Subject: 2nd message
Date: Wed, 11 May 2017 14:31:59 +0000
Message-ID: <0000001@localhost/>
Content-Type: text/plain

Hi there :)`
	err = mb.CreateMessage([]string{"\\Seen"}, time.Now(), imap.Literal(bytes.NewBufferString(msg)))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new server
	s := server.New(be)
	s.Addr = ":1143"
	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Print("Starting IMAP server at localhost:1143")
	go func() {
		if err := s.ListenAndServe(); err != nil {
			if strings.Contains(err.Error(), "use of closed") {
				log.Lvl2("Quitting server")
				return
			}
			log.Fatal(err)
		}
	}()
	return s
}

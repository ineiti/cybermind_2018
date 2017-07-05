package user

import (
	"path/filepath"

	"syscall"

	"os"

	"bufio"

	"strings"
	"sync"

	"encoding/json"

	"fmt"

	"io/ioutil"

	"github.com/ineiti/cybermind/broker"
	"gopkg.in/dedis/onet.v1/log"
)

/*
CLI offers a simple command-line interface to spawn new modules and to
enter messages.
*/

const CLIName = "cli.fifo"

type CLI struct {
	broker   *broker.Broker
	counter  int
	closed   chan bool
	fifoName string
	sync.Mutex
}

func RegisterCLI(b *broker.Broker) error {
	err := b.RegisterModule("cli", NewCLI)
	if err != nil {
		return err
	}
	return b.SpawnModule("cli", nil)
}

func NewCLI(b *broker.Broker, config []byte) broker.Module {
	c := &CLI{
		broker:   b,
		closed:   make(chan bool),
		fifoName: filepath.Join(broker.ConfigPath, CLIName),
	}
	if config != nil {
		err := json.Unmarshal(config, c)
		if err != nil {
			log.Error(err)
		}
	}
	log.ErrFatal(syscall.Mkfifo(c.fifoName, 0600))
	go c.Listen()
	return c
}

func (c *CLI) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	if m == nil {
		return nil, nil
	}
	for _, a := range m.Actions {
		if a.Command == "Stop" {
			c.Stop()
		}
	}
	return nil, nil
}

func (c *CLI) Listen() {
	defer func() {
		log.LLvl3("Finishing listen")
		c.closed <- true
	}()
	for {
		log.Print("Waiting on", c.fifoName)
		fifoFile, err := os.OpenFile(c.fifoName, os.O_RDONLY, 0600)
		if err != nil {
			log.ErrFatal(err)
			continue
		}
		cmdStr, _, err := bufio.NewReader(fifoFile).ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			log.Error(err)
			return
		}
		c.Lock()
		args := strings.SplitN(string(cmdStr), "::", 2)
		if len(args) < 2 {
			continue
		}
		client := args[0]
		args = strings.SplitN(args[1], ":", 2)
		args = append(args, "")
		cmd, arg := args[0], args[1]
		log.LLvlf3("%s sent cmd: %s with args: %s", client, cmd, arg)
		switch cmd {
		case "list":
			log.LLvl3("Listing")
			var names []string
			for m := range c.broker.ModuleTypes {
				names = append(names, m)
			}
			out := fmt.Sprintf(strings.Join(names, ",") + "\n")
			err := ioutil.WriteFile(client, []byte(out), 0600)
			if err != nil {
				log.Error(err)
			}
		case "close":
			log.LLvl3("Closing")
			return
		default:
			log.Warn("Didn't recognize command", cmd, arg)
		}
		c.Unlock()
		fifoFile.Close()
	}
}

func (c *CLI) Stop() {
	err := ioutil.WriteFile(c.fifoName, []byte("0::close\n"), 0600)
	log.ErrFatal(err)
	<-c.closed
	log.ErrFatal(os.RemoveAll(c.fifoName))
}

func CLICmd(path, cmd string) (reply string, err error) {
	repName := broker.GetTemp("fifo")
	if err = syscall.Mkfifo(repName, 0600); err != nil {
		return
	}

	fullCmd := fmt.Sprintf("%s::%s\n", repName, cmd)
	fifoName := filepath.Join(path, CLIName)
	log.Print("Writing to", fifoName)
	err = ioutil.WriteFile(fifoName, []byte(fullCmd), 0600)
	if err != nil {
		return
	}

	repFile, err := os.OpenFile(repName, os.O_RDONLY, 0600)
	if err != nil {
		return
	}
	rep := bufio.NewReader(repFile)
	reply, err = rep.ReadString('\n')
	if err != nil {
		return
	}
	reply = strings.TrimSpace(reply)
	return
}

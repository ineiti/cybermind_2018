package broker

import (
	"strings"

	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/dedis/crypto/random"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/dedis/onet.v1/log"
	"os"
)

const BaseDomain = "cybermind.gasser.blue"

var ConfigPath string

func SubDomain(dom ...string) string {
	return strings.Join(append(dom, BaseDomain), ".")
}

func init() {
	ResetConfigPath()
}

func ResetConfigPath() {
	// Create a somewhat reasonable default-config-path
	switch runtime.GOOS {
	case "darwin":
		ConfigPath = "Library/CyberMind"
	case "linux":
		ConfigPath = ".config/cybermind"
	default:
		ConfigPath = "CyberMind"
		log.Warn("sorry - this platform is not yet supported - please contact ineiti@gasser.blue")
	}

	d, err := homedir.Dir()
	log.ErrFatal(err)
	ConfigPath = filepath.Join(d, ConfigPath)
}

func TempConfigPath() (err error) {
	ConfigPath, err = ioutil.TempDir("", "CyberMind")
	return err
}

func GetTemp() string {
	f, err := ioutil.TempFile("", "config.toml")
	log.ErrFatal(err)
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}

// ObjectID represents any object
type ObjectID []byte
type ModuleID []byte
type TagID []byte
type ActionID []byte
type MessageID []byte

type Tag struct {
	Object    ObjectID
	Key       string
	Value     string
	Ephemeral bool
	ID        TagID
}

type Object struct {
	Module ModuleID
	ID     ObjectID
}

type KeyValue struct {
	Key   string
	Value string
}

type Action struct {
	Command   string
	Arguments map[string]string
	// TTL indicates how many hops this action should take. TTL == 0
	// means only local propagation.
	TTL int
	ID  ActionID
}

func NewAction(cmd string, kvs ...KeyValue) Action {
	args := map[string]string{}
	for _, kv := range kvs {
		args[kv.Key] = kv.Value
	}
	return Action{
		Command:   cmd,
		Arguments: args,
		ID:        ActionID(random.Bytes(32, random.Stream)),
	}
}

type Message struct {
	Objects []Object
	Tags    []Tag
	Actions []Action
	ID      MessageID
}

type Module interface {
	ProcessMessage(m *Message) ([]Message, error)
}

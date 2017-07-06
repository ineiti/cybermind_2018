package broker

import (
	"strings"

	"io/ioutil"
	"path/filepath"
	"runtime"

	"crypto/sha256"
	"os"

	"sort"

	"encoding/binary"

	"fmt"

	"github.com/dedis/crypto/random"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/dedis/onet.v1/log"
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

func GetTemp(name string) string {
	f, err := ioutil.TempFile("", name)
	log.ErrFatal(err)
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}

// ObjectID represents any object
const IDLen = 32

type ObjectID []byte
type ModuleID []byte
type TagID []byte
type ActionID []byte
type MessageID []byte

func NewObjectID() ObjectID {
	return random.Bytes(IDLen, random.Stream)
}
func NewModuleID() ModuleID {
	return random.Bytes(IDLen, random.Stream)
}
func NewTagID() TagID {
	return random.Bytes(IDLen, random.Stream)
}
func NewActionID() ActionID {
	return random.Bytes(IDLen, random.Stream)
}
func NewMessageID() MessageID {
	return random.Bytes(IDLen, random.Stream)
}

type Tag struct {
	gorm.Model
	Key       string
	Value     string
	Ephemeral bool
	Objects   []Object `gorm:"many2many:ObjectTag;"`
	GID       TagID
}

func NewTag(key, value string) Tag {
	return Tag{
		Key:   key,
		Value: value,
		GID:   NewTagID(),
	}
}

func (t *Tag) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(t.Key))
	hash.Write([]byte(t.Value))
	if t.Ephemeral {
		hash.Write([]byte{1})
	} else {
		hash.Write([]byte{0})
	}
	hash.Write(t.GID)
	return hash.Sum(nil)
}

func (t Tag) String() string {
	return fmt.Sprintf(">%x:%s:%s - %t/%x - %s<", t.ID, t.Key, t.Value, t.Ephemeral,
		t.GID[0:4], t.Objects)
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
	GID ActionID
}

func NewAction(cmd string, kvs ...KeyValue) Action {
	args := map[string]string{}
	for _, kv := range kvs {
		args[kv.Key] = kv.Value
	}
	return Action{
		Command:   cmd,
		Arguments: args,
		GID:       NewActionID(),
	}
}

func (a *Action) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(a.Command))
	binary.Write(hash, binary.BigEndian, a.TTL)
	hash.Write(a.GID)
	var args []string
	for arg := range a.Arguments {
		args = append(args, arg)
	}
	sort.Strings(args)
	for arg, value := range args {
		binary.Write(hash, binary.BigEndian, arg)
		hash.Write([]byte(value))
	}
	return hash.Sum(nil)
}

type Tags []Tag

func (t Tags) GetLastValue(key string) *Tag {
	var last *Tag
	for _, tag := range t {
		if tag.Key == key {
			last = &tag
		}
	}
	return last
}

func (t Tags) Add(tag Tag) Tags {
	return append(t, tag)
}

type Message struct {
	Objects   []Object
	Tags      Tags
	Action    Action
	ID        MessageID
	Signature []byte
}

func (m *Message) Hash() []byte {
	hash := sha256.New()
	for _, o := range m.Objects {
		hash.Write(o.Hash())
	}
	for _, t := range m.Tags {
		hash.Write(t.Hash())
	}
	hash.Write(m.Action.Hash())
	hash.Write(m.ID)
	return hash.Sum(nil)
}

func (m *Message) String() string {
	var str []string
	var objs []string
	for _, o := range m.Objects {
		objs = append(objs, fmt.Sprintf("ID: %x - ModuleID: %x",
			o.GID[0:4], o.ModuleID[0:4]))
	}
	if len(objs) > 0 {
		str = append(str, fmt.Sprintf("objects: [%s]", strings.Join(objs, ",")))
	}
	var tags []string
	for _, t := range m.Tags {
		tags = append(tags, fmt.Sprintf("{'%s':%s}", t.Key, t.Value))
	}
	if len(tags) > 0 {
		str = append(str, fmt.Sprintf("tags: [%s]", strings.Join(tags, ",")))
	}
	if m.Action.Command != "" {
		str = append(str, fmt.Sprintf("action: %+v", m.Action))
	}
	return strings.Join(str, " - ")
}

type Module interface {
	ProcessMessage(m *Message) ([]Message, error)
}

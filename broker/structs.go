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

	"unicode"

	"github.com/dedis/crypto/random"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/dedis/onet.v1/log"
)

const BaseDomain = "cybermind.gasser.blue"

var ConfigPath string

func init() {
	ResetConfigPath()
}

// ObjectID represents any object
const IDLen = 32

type ModuleID []byte
type MessageID []byte
type ObjectID []byte
type TagID []byte
type ActionID []byte
type StatusID []byte

func NewModuleID() ModuleID {
	return random.Bytes(IDLen, random.Stream)
}
func NewMessageID() MessageID {
	return random.Bytes(IDLen, random.Stream)
}
func NewObjectID() ObjectID {
	return random.Bytes(IDLen, random.Stream)
}
func NewTagID() uint64 {
	return random.Uint64(random.Stream) / 2
}
func NewActionID() ActionID {
	return random.Bytes(IDLen, random.Stream)
}

type Module interface {
	ProcessMessage(m *Message) ([]Message, error)
}

type Message struct {
	ID        MessageID
	Objects   []Object
	Tags      Tags
	Action    Action
	Status    Status
	Signature []byte
}

func NewMessage(objs *[]Object, tags *Tags, action *Action, status *Status) Message {
	return Message{
		ID:      NewMessageID(),
		Objects: *objs,
		Tags:    *tags,
		Action:  *action,
		Status:  *status,
	}
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
	str := []string{fmt.Sprintf("%x", m.ID[0:4])}
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
		tags = append(tags, fmt.Sprintf("{'%s':%+q}", t.Key, t.Value))
	}
	if len(tags) > 0 {
		str = append(str, fmt.Sprintf("tags: [%s]", strings.Join(tags, ",")))
	}
	if m.Action.Command != "" {
		str = append(str, fmt.Sprintf("action: %+v", m.Action))
	}
	return strings.Join(str, " - ")
}

type Object struct {
	gorm.Model
	GID        ObjectID
	ModuleID   ModuleID
	Data       []byte
	Tags       Tags `gorm:"many2many:ObjectTag;"`
	IgnoreData bool
}

func NewObject(id ModuleID, data []byte) *Object {
	return &Object{
		ModuleID: id,
		GID:      random.Bytes(IDLen, random.Stream),
		Data:     data,
	}
}

func (o *Object) Hash() []byte {
	hash := sha256.New()
	hash.Write(o.GID)
	hash.Write(o.ModuleID)
	hash.Write(o.Data)
	return hash.Sum(nil)
}

func (o Object) String() string {
	dataLen := len(o.Data)
	if dataLen > 4 {
		dataLen = 4
	}
	return fmt.Sprintf(">%x:%x/%x - %t/%x - %s<", o.ID, o.ModuleID[0:4], o.Data[0:dataLen],
		o.IgnoreData, o.GID[0:4], o.Tags)
}

func (o Object) Copy() Object {
	return Object{
		GID:        o.GID,
		ModuleID:   o.ModuleID,
		Data:       o.Data,
		Tags:       o.Tags,
		IgnoreData: o.IgnoreData,
	}
}

type Tag struct {
	ID        uint64
	GID       uint64
	Key       string
	Value     string
	Ephemeral bool
	Objects   []Object         `gorm:"many2many:ObjectTag;"`
	TagA      []TagAssociation `gorm:"many2many:TagIsA;"`
}

type IsAssociation uint64

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
	//hash.Write(t.GID)
	return hash.Sum(nil)
}

func (t Tag) String() string {
	var tags []string
	for _, tag := range t.TagA {
		tags = append(tags, fmt.Sprintf("%x:%s", tag.Association, tag.Tags))
		//tags = append(tags, fmt.Sprintf("%x", tag.GID[0:4]))
	}
	var objs []string
	for _, obj := range t.Objects {
		objs = append(objs, fmt.Sprintf("%x", obj.GID[0:4]))
	}
	//return fmt.Sprintf("<<%x:[%t]%s=%s-%s/%s>>", t.GID[0:4], t.Ephemeral,
	return fmt.Sprintf("<<%x:[%t]%s=%s-%s/%s>>", t.GID, t.Ephemeral,
		t.Key, toString(t.Value), objs, tags)
}

// AddAssociation adds one or more tags with a given association.
func (t *Tag) AddAssociation(tags Tags, asso IsAssociation) TagAssociation {
	tia := TagAssociation{
		Association: uint64(asso),
		Tags:        tags,
	}
	t.TagA = append(t.TagA, tia)
	return tia
}

type Tags []Tag

func (t Tags) GetLastValue(key string) *Tag {
	var last *Tag
	for i, tag := range t {
		if tag.Key == key {
			last = &t[i]
		}
	}
	return last
}

func (t Tags) Add(tag Tag) Tags {
	return append(t, tag)
}

const (
	AssociationIsA IsAssociation = iota
	AssociationEquals
	AssociationTranslation
	AssociationAbbreviation
)

type TagAssociation struct {
	ID          uint64
	Association uint64
	Tags        Tags `gorm:"many2many:TagIsA;"`
}

type Action struct {
	Command   string
	Arguments KeyValues
	// TTL indicates how many hops this action should take. TTL == 0
	// means only local propagation.
	TTL int
}

func NewAction(cmd string, kvs ...KeyValue) Action {
	args := KeyValues{}
	for _, kv := range kvs {
		args[kv.Key] = kv.Value
	}
	return Action{
		Command:   cmd,
		Arguments: args,
	}
}

func (a *Action) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(a.Command))
	binary.Write(hash, binary.BigEndian, a.TTL)
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

func (a *Action) String() string {
	return fmt.Sprintf("[%s:%+q]", a.Command, a.Arguments)
}

type Status struct {
	Entries KeyValues
}

func NewStatus(kv ...KeyValue) Status {
	s := Status{
		Entries: KeyValues{},
	}
	for _, e := range kv {
		s.Entries[e.Key] = e.Value
	}
	return s
}

func SubDomain(dom ...string) string {
	return strings.Join(append(dom, BaseDomain), ".")
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

type KeyValue struct {
	Key   string
	Value string
}

func NewKeyValue(key, value string) KeyValue {
	return KeyValue{Key: key, Value: value}
}

func (kv KeyValue) String() string {
	return fmt.Sprintf("%s=%s", kv.Key, toString(kv.Value))
}

type KeyValues map[string]string

func (kvs KeyValues) String() string {
	var ret []string
	for k, v := range kvs {
		ret = append(ret, fmt.Sprintf("%s=%s", k, toString(v)))
	}
	return strings.Join(ret, ",")
}

func toString(s string) string {
	for _, c := range s {
		if !unicode.IsPrint(c) {
			return fmt.Sprintf("%x", []byte(s))
		}
	}
	return s
}

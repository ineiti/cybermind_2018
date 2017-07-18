package base

import (
	"path/filepath"

	"fmt"

	"regexp"

	"encoding/hex"

	"github.com/ineiti/cybermind/broker"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gopkg.in/dedis/onet.v1/log"
)

var StorageSearchTag = broker.SubDomain("search_tag", "storage")
var StorageSearchObject = broker.SubDomain("search_obj", "storage")

var ModuleStorage = "storage"

const StorageDB = "cybermind.db"

type Storage struct {
	Broker   *broker.Broker
	DataBase *gorm.DB
}

func RegisterStorage(b *broker.Broker) error {
	err := b.RegisterModule(ModuleStorage, NewStorage)
	if err != nil {
		return err
	}
	_, err = b.SpawnModule(ModuleStorage, nil)
	return err
}

func NewStorage(b *broker.Broker, msg *broker.Message) broker.Module {
	s := &Storage{
		Broker: b,
	}
	var err error
	path := filepath.Join(broker.ConfigPath, StorageDB)
	fmt.Println("echo .dump | sqlite3", path)
	s.DataBase, err = gorm.Open("sqlite3", path)
	log.ErrFatal(err)
	log.ErrFatal(s.DataBase.AutoMigrate(&broker.Tags{}).Error)
	log.ErrFatal(s.DataBase.AutoMigrate(&broker.Object{}).Error)
	log.ErrFatal(s.DataBase.AutoMigrate(&broker.TagAssociation{}).Error)
	log.Lvl2("Initialized database")
	return s
}

func (s *Storage) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	log.Lvlf3("Message is: %+v", m)
	if m == nil {
		return nil, nil
	}
	switch m.Action.Command {
	case broker.BrokerStop:
		log.Lvl2("Closing database")
		if err := s.DataBase.Close(); err != nil {
			log.Error(err)
		}
		return nil, nil
	case StorageSearchTag:
		log.Lvl3("Searching", m.Action.Arguments)
		find := createFind(m.Action.Arguments)
		var tags []broker.Tag
		err := s.DataBase.Where(find).Find(&tags).Error
		if err != nil {
			log.Error(err)
		}
		for i := range tags {
			s.DataBase.Model(&tags[i]).Related(&tags[i].Objects, "Objects")
			s.DataBase.Model(&tags[i]).Related(&tags[i].TagA, "TagA")
			// We arbitrarily decide that two levels of 'Related' are
			// enough - if you need more, do another `StorageSearchTag`
			// call.
			for j := range tags[i].TagA {
				s.DataBase.Model(&tags[i].TagA).Related(&tags[i].TagA[j].Tags, "Tags")
			}
		}
		if len(tags) == 0 {
			log.Lvl3("Didn't find any tags for", find)
		}
		log.Lvl3("Found Tags:", tags)
		return []broker.Message{{
			Tags: tags,
		}}, nil
	case StorageSearchObject:
		find := createFind(m.Action.Arguments)
		var objs []broker.Object
		err := s.DataBase.Where(find).Find(&objs).Error
		if err != nil {
			log.Error(err)
			return nil, nil
		}
		for i := range objs {
			s.DataBase.Model(&objs[i]).Related(&objs[i].Tags, "Tags")
		}
		if len(objs) == 0 {
			log.Lvlf3("Didn't find any objects for %#v", find)
			return nil, nil
		}
		log.Lvl3("Found Objs:", objs)
		return []broker.Message{{
			ID:      broker.NewMessageID(),
			Objects: objs,
		}}, nil
	default:
	}
	for i, o := range m.Objects {
		obj := o.Copy()
		if o.IgnoreData {
			obj.Data = []byte{}
		}
		if o.ID > 0 {
			log.Lvl2("Not storing already indexed object")
			continue
		}
		err := s.DataBase.Create(&m.Objects[i]).Error
		if err != nil {
			return nil, err
		}
		log.Lvl3("Stored Object:", o)
	}
	for _, t := range m.Tags {
		if !t.Ephemeral {
			if t.ID > 0 {
				log.Lvl2("Not storing already indexed tag")
				continue
			}
			for _, o := range m.Objects {
				t.Objects = append(t.Objects, o)
			}
			err := s.DataBase.Create(&t).Error
			if err != nil {
				return nil, err
			}
			log.Lvl3("Stored Tag:", t, t.Objects)
		}
		//printTags(s)
		//printObjects(s)
	}
	return nil, nil
}

func printTags(s *Storage) {
	var tags []broker.Tag
	s.DataBase.Find(&tags)
	log.Print(tags)
}

func printObjects(s *Storage) {
	var objs []broker.Object
	s.DataBase.Find(&objs)
	log.Print(objs)
}

func createFind(args map[string]string) map[string]interface{} {
	find := make(map[string]interface{})
	for k, v := range args {
		switch k {
		default:
			log.Lvlf4("Adding search-argument %s/%s", k, v)
			find[k] = interpretString(v)
		}
	}
	return find
}

func interpretString(str string) interface{} {
	if b, _ := regexp.MatchString("X'[0-9a-f]*'", str); b {
		bin, err := hex.DecodeString(str[2 : len(str)-1])
		if err == nil {
			log.Lvlf4("Converted to binary %#v", bin)
			return bin
		} else {
			log.Error(err)
		}
	}
	return str
}

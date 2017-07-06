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

var StorageSave = broker.NewTag(broker.SubDomain("save", "storage"), "true")
var StorageSaveTags = broker.Tags{StorageSave}

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
	return b.SpawnModule(ModuleStorage, nil)
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
	//log.ErrFatal(s.DataBase.Model(&broker.Tag{}).
	//	Related(&broker.Object{}, "Objects").Error)
	return s
}

func (s *Storage) ProcessMessage(m *broker.Message) ([]broker.Message, error) {
	switch m.Action.Command {
	case broker.BrokerStop:
		log.Print("Closing database")
		if err := s.DataBase.Close(); err != nil {
			log.Error(err)
		}
		return nil, nil
	case StorageSearchTag:
		log.Print("Search")
		find := createFind(m.Action.Arguments)
		var tags []broker.Tag
		err := s.DataBase.Where(find).Find(&tags).Error
		if err != nil {
			log.Error(err)
		}
		for i := range tags {
			s.DataBase.Model(&tags[i]).Related(&tags[i].Objects, "Objects")
		}
		log.Print("Found Tags:", tags)
		return []broker.Message{{
			Tags: tags,
		}}, nil
	case StorageSearchObject:
		find := createFind(m.Action.Arguments)
		var objs []broker.Object
		err := s.DataBase.Where(find).Find(&objs).Error
		if err != nil {
			log.Error(err)
		}
		for i := range objs {
			s.DataBase.Model(&objs[i]).Related(&objs[i].Tags, "Tags")
		}
		log.Print("Found Objs:", objs)
		return []broker.Message{{
			Objects: objs,
		}}, nil
	default:
	}
	for i, o := range m.Objects {
		if o.StoreData {
			log.Print("Storing Object:", o)
			o.StoreData = false
			err := s.DataBase.Create(&m.Objects[i]).Error
			if err != nil {
				return nil, err
			}
		}
		printObjects(s)
	}
	for _, t := range m.Tags {
		if !t.Ephemeral {
			for _, o := range m.Objects {
				t.Objects = append(t.Objects, o)
			}
			log.Print("Storing Tag:", t, t.Objects)
			err := s.DataBase.Create(&t).Error
			if err != nil {
				return nil, err
			}
		}
		printTags(s)
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
			log.Printf("Adding search-argument %s/%s", k, v)
			find[k] = interpretString(v)
		}
	}
	return find
}

func interpretString(str string) interface{} {
	if b, _ := regexp.MatchString("X'[0-9a-f]*'", str); b {
		bin, err := hex.DecodeString(str[2 : len(str)-1])
		if err == nil {
			log.Printf("Converted to binary %x", bin)
			return bin
		} else {
			log.Error(err)
		}
	}
	return str
}

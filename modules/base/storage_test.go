package base

import (
	"testing"

	"path/filepath"

	"fmt"

	"regexp"

	"os"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/test"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"gopkg.in/dedis/onet.v1/log"
)

type TestStruct struct {
	gorm.Model
	Key   string
	Value string
}

func TestReg(t *testing.T) {
	log.Print(regexp.MatchString("X'[0-9a-f]*'", "X'1234'"))
}

func TestGorm(t *testing.T) {
	path := filepath.Join(broker.ConfigPath, StorageDB)
	db, err := gorm.Open("sqlite3", path)
	log.ErrFatal(err)
	log.ErrFatal(db.AutoMigrate(&broker.Tags{}).Error)
	log.ErrFatal(db.AutoMigrate(&broker.Object{}).Error)
	obj := broker.Object{
		GID:       broker.NewObjectID(),
		ModuleID:  ModuleIDConfig,
		StoreData: true,
	}
	obj2 := broker.Object{
		GID:       broker.NewObjectID(),
		ModuleID:  ModuleIDConfig,
		StoreData: true,
	}
	log.ErrFatal(db.Create(&obj).Error)

	tag := broker.Tag{
		GID:     broker.NewTagID(),
		Objects: []broker.Object{obj, obj2},
	}
	log.Print(tag)
	log.ErrFatal(db.Create(&tag).Error)
	var objs []broker.Object
	db.Find(&objs)
	log.Print(objs)
	log.Print(objs[0].ModuleID)

	find := make(map[string]interface{})
	find["module_id"] = []byte{0, 0, 0, 0}
	db.LogMode(true)

	db.Where(find).Find(&objs)
	log.Print(objs)

	var tags []broker.Tag
	db.Find(&tags)
	db.Model(&tags[0]).Related(&tags[0].Objects, "Objects")
	log.Print(tags)

}

func TestStorageSave(t *testing.T) {
	sb := initStorageBroker(true)
	sb.Broker.BroadcastMessage(&broker.Message{
		Objects: []broker.Object{{
			GID:       broker.NewObjectID(),
			ModuleID:  ModuleIDConfig,
			StoreData: true,
		}},
		Tags: broker.Tags{broker.NewTag("test", "123")},
	})
	sb.Broker.Stop()

	log.Lvl1("Starting new broker")
	sb = initStorageBroker(false)
	sb.Broker.BroadcastMessage(&broker.Message{
		Action: broker.Action{
			Command: StorageSearchObject,
			Arguments: map[string]string{
				"module_id": fmt.Sprintf("X'%x'", ModuleIDConfig),
			},
		},
	})
	log.Print(sb.Logger.Messages)
	require.Equal(t, 2, len(sb.Logger.Messages))
	sb.Broker.Stop()
}

func TestStorageSaveRelation(t *testing.T) {
	sb := initStorageBroker(true)
	objs := []broker.Object{
		{
			GID:       broker.NewObjectID(),
			ModuleID:  ModuleIDConfig,
			StoreData: true,
			Data:      []byte("one"),
		},
		{
			GID:       broker.NewObjectID(),
			ModuleID:  ModuleIDConfig,
			StoreData: true,
			Data:      []byte("two"),
		},
	}
	tags := broker.Tags{broker.NewTag("test", "123"),
		broker.NewTag("test2", "456")}

	sb.Broker.BroadcastMessage(&broker.Message{
		Objects: objs,
		Tags:    tags,
	})
	sb.Broker.Stop()

	log.Lvl1("Starting new broker")
	sb = initStorageBroker(false)
	sb.Broker.BroadcastMessage(&broker.Message{
		Action: broker.Action{
			Command: StorageSearchTag,
			Arguments: map[string]string{
				"key": "test",
			},
		},
	})
	log.Print(sb.Logger.Messages)
	//sb.Broker.BroadcastMessage(&broker.Message{
	//	Action: broker.Action{
	//		Command: StorageSearchTag,
	//		Arguments: map[string]string{
	//			"object_gid": fmt.Sprintf("X'%x'", sb.Logger.Messages[1].Objects[0].GID),
	//		},
	//	},
	//})
	require.Equal(t, 2, len(sb.Logger.Messages))
	sb.Broker.Stop()
}

type storBrok struct {
	Broker *broker.Broker
	Logger *test.Logger
}

func initStorageBroker(clear bool) *storBrok {
	if clear {
		path := filepath.Join(broker.ConfigPath, StorageDB)
		os.Remove(path)
	}
	sb := &storBrok{
		Broker: broker.NewBroker(),
	}
	RegisterStorage(sb.Broker)
	sb.Logger = test.SpawnLogger(sb.Broker)
	return sb
}

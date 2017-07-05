package base

import (
	"testing"

	"path/filepath"

	"fmt"

	"regexp"

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
	obj := &broker.Object{
		GID:      broker.NewObjectID(),
		ModuleID: ModuleIDConfig,
		StoreIt:  true,
	}
	log.ErrFatal(db.Create(&obj).Error)
	var objs []broker.Object
	db.Find(&objs)
	log.Print(objs)
	log.Print(objs[0].ModuleID)

	find := make(map[string]interface{})
	find["module_id"] = []byte{0, 0, 0, 0}
	db.LogMode(true)

	db.Where(find).Find(&objs)
	log.Print(objs)
}

func TestStorageSave(t *testing.T) {
	sb := initStorageBroker()
	sb.Broker.BroadcastMessage(&broker.Message{
		Objects: []broker.Object{{
			GID:      broker.NewObjectID(),
			ModuleID: ModuleIDConfig,
			StoreIt:  true,
		}},
		Tags: broker.Tags{broker.NewTag("test", "123")},
	})
	sb.Broker.Stop()

	log.Lvl1("Starting new broker")
	sb = initStorageBroker()
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

type storBrok struct {
	Broker *broker.Broker
	Logger *test.Logger
}

func initStorageBroker() *storBrok {
	sb := &storBrok{
		Broker: broker.NewBroker(),
	}
	RegisterStorage(sb.Broker)
	sb.Logger = test.SpawnLogger(sb.Broker)
	return sb
}

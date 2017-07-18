package base

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/dedis/onet/log"
	"github.com/ineiti/cybermind/broker"
	"github.com/jinzhu/gorm"
)

type TestStruct struct {
	ID    uint64
	Key   string
	Value string
	TSIA  []TestStructIsA `gorm:"many2many:tstsia"`
}

type TestStructIsA struct {
	ID          uint64
	Association uint
	TS          []TestStruct `gorm:"many2many:tstsia"`
}

func TestGorm2(t *testing.T) {
	t.Skip("Gorm-test")
	path := filepath.Join(broker.ConfigPath, StorageDB)
	db, err := gorm.Open("sqlite3", path)
	log.ErrFatal(err)
	db.LogMode(true)
	log.ErrFatal(db.AutoMigrate(&TestStruct{}).Error)
	log.ErrFatal(db.AutoMigrate(&TestStructIsA{}).Error)
	tsia := &TestStructIsA{
		TS: []TestStruct{mts("count", "one"),
			mts("count", "two")},
	}
	log.ErrFatal(db.Create(&tsia).Error)

	tsia2 := &TestStructIsA{
		TS: []TestStruct{tsia.TS[1],
			mts("count", "three")},
	}
	log.ErrFatal(db.Create(&tsia2).Error)

	var tss []TestStruct
	log.ErrFatal(db.Find(&tss).Error)
	for i := range tss {
		for j := range tss[i].TSIA {
			db.Model(&tss[i].TSIA[j]).Related(&tss[i].TSIA[j].TS, "TS")
		}
	}
	log.Print(tss)

	var tsias []TestStructIsA
	log.ErrFatal(db.Find(&tsias).Error)
	db.Model(&tsias[0]).Related(&tsias[0].TS, "TS")
	db.Model(&tsias[1]).Related(&tsias[1].TS, "TS")
	//db.Model(&tags[0]).Related(&tags[0].Objects, "Objects")
	log.Print(tsias)
}

func TestReg(t *testing.T) {
	t.Skip("regexp-test")
	log.Print(regexp.MatchString("X'[0-9a-f]*'", "X'1234'"))
}

func TestGorm(t *testing.T) {
	//t.Skip("gorm-test")
	path := filepath.Join(broker.ConfigPath, StorageDB)
	db, err := gorm.Open("sqlite3", path)
	log.ErrFatal(err)
	log.ErrFatal(db.AutoMigrate(&broker.Tags{}).Error)
	log.ErrFatal(db.AutoMigrate(&broker.Object{}).Error)
	objs := getObjs(2)
	log.ErrFatal(db.Create(&objs[0]).Error)

	tag := broker.Tag{
		GID:     broker.NewTagID(),
		Objects: []broker.Object{objs[0], objs[1]},
	}
	log.Print(tag)
	log.ErrFatal(db.Create(&tag).Error)
	var objsRead []broker.Object
	db.Find(&objsRead)
	log.Print(objsRead)
	log.Print(objsRead[0].ModuleID)

	find := make(map[string]interface{})
	find["module_id"] = []byte{0, 0, 0, 0}
	db.LogMode(true)

	db.Where(find).Find(&objsRead)
	log.Print(objsRead)

	var tags []broker.Tag
	db.Find(&tags)
	db.Model(&tags[0]).Related(&tags[0].Objects, "Objects")
	log.Print(tags)
	db.Close()
}

func mts(k, v string) TestStruct {
	return TestStruct{Key: k, Value: v}
}

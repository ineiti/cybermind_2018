package broker

import (
	"crypto/sha256"
	"fmt"

	"github.com/jinzhu/gorm"
)

type Object struct {
	gorm.Model
	GID       ObjectID
	ModuleID  ModuleID
	Data      []byte
	Tags      []Tag `gorm:"many2many:ObjectTag;"`
	StoreData bool
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
		o.StoreData, o.GID[0:4], o.Tags)
}

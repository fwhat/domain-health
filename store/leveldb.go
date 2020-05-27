package store

import (
	"fmt"
	"github.com/Dowte/domain-health/common"
	"github.com/Dowte/domain-health/config"
	"github.com/Dowte/domain-health/pkg/file"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

var log = common.Log

var domainStore *DomainStore

func InitDomainStore() {
	dir := config.Instance.StoreDir
	if !file.PathExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			panic(fmt.Sprintf("%v: %s", err, dir))
		}
	}

	db, err := leveldb.OpenFile(dir, nil)

	if err != nil {
		panic(err)
	}

	domainStore = &DomainStore{
		db: db,
	}
}

func GetDomainStore() *DomainStore {
	return domainStore
}

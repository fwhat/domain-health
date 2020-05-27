package store

import (
	"fmt"
	"github.com/Dowte/domain-health/store/model"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type DomainStore struct {
	db *leveldb.DB
}

func (d *DomainStore) ReadAllDomainList() []*model.Domain {
	list := make([]*model.Domain, 0)
	iterator := d.db.NewIterator(util.BytesPrefix([]byte("domain/domainList/")), nil)
	defer iterator.Release()

	for iterator.Next() {
		item := &model.Domain{}
		err := item.UnmarshalFrom(iterator.Value())
		if err == nil {
			list = append(list, item)
		} else {
			log.Errorf("unmarshal domain err: %v", err)
		}
	}

	return list
}

func (d *DomainStore) getKeyByUuid(uuid string) []byte {
	return []byte(fmt.Sprintf("domain/domainList/%s", uuid))
}

func (d *DomainStore) SaveDomainInfo(item *model.Domain) bool {
	return d.db.Put(d.getKeyByUuid(item.Uuid), item.Marshal(), nil) == nil
}

func (d *DomainStore) DeleteDomainByUuid(uuid string) bool {
	return d.db.Delete(d.getKeyByUuid(uuid), nil) == nil
}

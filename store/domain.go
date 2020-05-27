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

func (d *DomainStore) ReadAllDomainByFrom(from model.From) []*model.Domain {
	list := make([]*model.Domain, 0)
	iterator := d.db.NewIterator(util.BytesPrefix([]byte("domain/domainList/")), nil)
	defer iterator.Release()

	for iterator.Next() {
		item := &model.Domain{}
		err := item.UnmarshalFrom(iterator.Value())
		if err == nil {
			if item.From == from {
				list = append(list, item)
			}
		} else {
			log.Errorf("unmarshal domain err: %v", err)
		}
	}

	return list
}

func (d *DomainStore) ReadAllDomainListNoError() []*model.Domain {
	list := make([]*model.Domain, 0)
	iterator := d.db.NewIterator(util.BytesPrefix([]byte("domain/domainList/")), nil)
	defer iterator.Release()

	for iterator.Next() {
		item := &model.Domain{}
		err := item.UnmarshalFrom(iterator.Value())
		if err == nil {
			if item.CheckError == "" {
				list = append(list, item)
			}
		} else {
			log.Errorf("unmarshal domain err: %v", err)
		}
	}

	return list
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

func (d *DomainStore) getKeyByAddress(address string) []byte {
	return []byte(fmt.Sprintf("domain/domainList/%s", address))
}

func (d *DomainStore) SaveDomainInfo(item *model.Domain) bool {
	return d.db.Put(d.getKeyByAddress(item.Address), item.Marshal(), nil) == nil
}

func (d *DomainStore) DeleteDomainByAddress(address string) bool {
	return d.db.Delete(d.getKeyByAddress(address), nil) == nil
}

func (d *DomainStore) HasDomainByAddress(address string) bool {
	has, err := d.db.Has(d.getKeyByAddress(address), nil)

	return has && err == nil
}

func (d *DomainStore) DeleteAllByFrom(from model.From) []*model.Domain {
	list := make([]*model.Domain, 0)
	iterator := d.db.NewIterator(util.BytesPrefix([]byte("domain/domainList/")), nil)
	defer iterator.Release()

	for iterator.Next() {
		item := &model.Domain{}
		err := item.UnmarshalFrom(iterator.Value())
		if err == nil {
			if item.From == from {
				d.db.Delete(iterator.Key(), nil)
			}
		} else {
			log.Errorf("unmarshal domain err: %v", err)
		}
	}

	return list
}

func (d *DomainStore) DeleteAddressArr(addressArr []string) {
	for _, address := range addressArr {
		d.db.Delete(d.getKeyByAddress(address), nil)
	}
}

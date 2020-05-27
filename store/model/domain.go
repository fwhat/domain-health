package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Domain struct {
	Uuid          string
	Address       string
	ExpireTime    time.Time
	LastCheckTime time.Time
	CheckError    string
}

func NewDomain() (d *Domain) {
	return &Domain{
		Uuid: uuid.New().String(),
	}
}

func NewDomainFromBytes(b []byte) (d *Domain) {
	d = &Domain{}
	err := d.UnmarshalFrom(b)
	if err != nil {
		return nil
	}

	return
}

func (d *Domain) UnmarshalFrom(b []byte) error {
	return json.Unmarshal(b, d)
}

func (d *Domain) Marshal() []byte {
	bytes, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	return bytes
}

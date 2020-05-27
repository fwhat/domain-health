package model

import (
	"encoding/json"
	"net/url"
	"time"
)

type From string

const (
	User   From = "user"
	Aliyun From = "aliyun"
)

type Domain struct {
	Record        *url.URL
	ExpireTime    time.Time
	LastCheckTime time.Time
	CheckError    string
	CreatedTime   time.Time
	From          From
}

func NewDomain() (d *Domain) {
	return &Domain{
		CreatedTime: time.Now(),
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

package subscriber

import "github.com/Dowte/domain-health/store/model"

type MessageType string

const (
	CretPreExpired MessageType = "CretPreExpired"
	CretExpired    MessageType = "CretExpired"
	ConnectTimeout MessageType = "ConnectTimeout"
)

type Message struct {
	Type   MessageType
	Domain *model.Domain
}

type Subscriber interface {
	AddMessage(Message)
	Delivery() error
}

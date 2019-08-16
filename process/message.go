package process

import "github.com/pborges/iot/pubsub"

type OnMessageFn func(Message) error

type Message struct {
	pubsub.MessageMetadata
	Process string
	Value   interface{}
}

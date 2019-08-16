package pubsub

type MessageMetadata struct {
	Id  string
	Key string
}

type Message struct {
	MessageMetadata
	Value interface{}
}

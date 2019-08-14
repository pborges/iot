package pubsub

type Message struct {
	Id    string
	Key   string
	Value interface{}
}

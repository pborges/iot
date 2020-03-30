package pubsub

type Definition interface {
	ValidateAndTransform(interface{}) (interface{}, error)
	Inspect(interface{}) string
	DefaultValue() interface{}
	Accept(interface{}) error
}

type Attribute struct {
	Name string
	Definition
}

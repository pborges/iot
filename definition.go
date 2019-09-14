package iot

type Definition interface {
	Transform(interface{}) (interface{}, error)
	DefaultValue() interface{}
}

package pubsub

import "reflect"

type StringDefinition struct {
	AcceptFn func(v string) error
}

func (d StringDefinition) ValidateAndTransform(v interface{}) (interface{}, error) {
	if v, ok := v.(string); ok {
		return v, nil
	}
	return nil, ErrInvalidType{Expected: reflect.String, Actual: reflect.TypeOf(v).Kind()}
}

func (d StringDefinition) Inspect(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (d StringDefinition) DefaultValue() interface{} {
	return ""
}
func (d StringDefinition) Accept(v interface{}) error {
	if d.AcceptFn != nil {
		return d.AcceptFn(v.(string))
	}
	return nil
}

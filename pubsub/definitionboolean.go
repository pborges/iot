package pubsub

import (
	"fmt"
	"reflect"
)

type BooleanDefinition struct {
	AcceptFn func(v bool) error
}

func (d BooleanDefinition) ValidateAndTransform(v interface{}) (interface{}, error) {
	var val bool
	switch i := v.(type) {
	case int:
		val = i > 0
		return val, nil
	case int8:
		val = i > 0
		return val, nil
	case int16:
		val = i > 0
		return val, nil
	case int32:
		val = i > 0
		return val, nil
	case int64:
		val = i > 0
		return val, nil
	case float32:
		val = i > 0
		return val, nil
	case float64:
		val = i > 0
		return val, nil
	case bool:
		return i, nil
	case string:
		return i == "true", nil
	}
	return nil, ErrInvalidType{Expected: reflect.Bool, Actual: reflect.TypeOf(v).Kind()}
}

func (d BooleanDefinition) Inspect(v interface{}) string {
	if s, ok := v.(bool); ok {
		return fmt.Sprintf("%t", s)
	}
	return ""
}

func (d BooleanDefinition) DefaultValue() interface{} {
	return false
}
func (d BooleanDefinition) Accept(v interface{}) error {
	if d.AcceptFn != nil {
		return d.AcceptFn(v.(bool))
	}
	return nil
}

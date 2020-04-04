package pubsub

import (
	"reflect"
	"strconv"
)

type IntegerDefinition struct {
	AcceptFn func(v int64) error
}

func (d IntegerDefinition) ValidateAndTransform(v interface{}) (interface{}, error) {
	var val int64
	switch i := v.(type) {
	case int:
		val = int64(i)
		return val, nil
	case int8:
		val = int64(i)
		return val, nil
	case int16:
		val = int64(i)
		return val, nil
	case int32:
		val = int64(i)
		return val, nil
	case int64:
		val = i
		return val, nil
	case float32:
		val = int64(i)
		return val, nil
	case float64:
		val = int64(i)
		return val, nil
	case string:
		return strconv.ParseInt(i, 10, 64)
	}
	return nil, ErrInvalidType{Expected: reflect.Int64, Actual: reflect.TypeOf(v).Kind()}
}

func (d IntegerDefinition) Inspect(v interface{}) string {
	if s, ok := v.(int64); ok {
		return strconv.FormatInt(s, 64)
	}
	return ""
}

func (d IntegerDefinition) DefaultValue() interface{} {
	return int64(0)
}
func (d IntegerDefinition) Accept(v interface{}) error {
	if d.AcceptFn != nil {
		return d.AcceptFn(v.(int64))
	}
	return nil
}

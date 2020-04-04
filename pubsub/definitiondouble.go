package pubsub

import (
	"reflect"
	"strconv"
)

type DoubleDefinition struct {
	AcceptFn func(v float64) error
}

func (d DoubleDefinition) ValidateAndTransform(v interface{}) (interface{}, error) {
	var val float64
	switch i := v.(type) {
	case int:
		val = float64(i)
		return val, nil
	case int8:
		val = float64(i)
		return val, nil
	case int16:
		val = float64(i)
		return val, nil
	case int32:
		val = float64(i)
		return val, nil
	case int64:
		val = float64(i)
		return val, nil
	case float32:
		val = float64(i)
		return val, nil
	case float64:
		val = i
		return val, nil
	case string:
		return strconv.ParseFloat(i, 64)
	}
	return nil, ErrInvalidType{Expected: reflect.Float64, Actual: reflect.TypeOf(v).Kind()}
}

func (d DoubleDefinition) Inspect(v interface{}) string {
	if s, ok := v.(float64); ok {
		return strconv.FormatFloat(s, 'f', 4, 64)
	}
	return ""
}

func (d DoubleDefinition) DefaultValue() interface{} {
	return 0.0
}
func (d DoubleDefinition) Accept(v interface{}) error {
	if d.AcceptFn != nil {
		return d.AcceptFn(v.(float64))
	}
	return nil
}

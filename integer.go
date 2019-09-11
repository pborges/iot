package iot

import "errors"

type IntegerDefinition struct {
	Min     int64
	Max     int64
	Default int64
}

func (a IntegerDefinition) DefaultValue() interface{} {
	return a.Default
}

func (a IntegerDefinition) Transform(value interface{}) (interface{}, error) {
	var data int64
	switch i := value.(type) {
	case int:
		data = int64(i)
	case int8:
		data = int64(i)
	case int16:
		data = int64(i)
	case int32:
		data = int64(i)
	case int64:
		data = i
	default:
		return nil, errors.New("unknown type")
	}
	if a.Min != 0 || a.Max != 0 {
		if data < a.Min {
			return nil, errors.New("value is less then min")
		}
		if data > a.Max {
			return nil, errors.New("value is more then max")
		}
	}
	return data, nil
}
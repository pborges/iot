package typed

import (
	"errors"
)

type IntegerAcceptor struct {
	Min int64
	Max int64
}

func (a IntegerAcceptor) Accept(value interface{}) error {
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
		return errors.New("unknown type")
	}
	if a.Min != 0 || a.Max != 0 {
		if data < a.Min {
			return errors.New("value is less then min")
		}
		if data > a.Max {
			return errors.New("value is more then max")
		}
	}
	return nil
}

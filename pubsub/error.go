package pubsub

import (
	"fmt"
	"reflect"
	"time"
)

type ErrNoValue struct {
	Attribute string
	Timestamp time.Time
}

func (e ErrNoValue) Error() string {
	return fmt.Sprintf("no value for attribute: '%s' at: %s", e.Attribute, e.Timestamp.Local().Format(time.RubyDate))
}

type ErrDuplicateAttribute struct {
	Attribute string
}

func (e ErrDuplicateAttribute) Error() string {
	return fmt.Sprintf("duplicate attribute '%s'", e.Attribute)
}

type ErrUnknownAttribute struct {
	Attribute string
}

func (e ErrUnknownAttribute) Error() string {
	return fmt.Sprintf("unknown attribute '%s'", e.Attribute)
}

type ErrInvalidType struct {
	Expected reflect.Kind
	Actual   reflect.Kind
}

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("invalid type expected '%s' but got '%s'", e.Expected, e.Actual)
}

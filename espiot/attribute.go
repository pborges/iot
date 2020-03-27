package espiot

import (
	"fmt"
	"time"
)

type AttributeAndValue interface {
	AttributeDef() Attribute
	InspectValue() string
}

type Attribute struct {
	Name      string
	ReadOnly  bool
	UpdatedAt time.Time
}

func (a Attribute) AttributeDef() Attribute {
	return a
}

type StringAttributeValue struct {
	Attribute
	Value string
}

func (a StringAttributeValue) InspectValue() string {
	return a.Value
}

type BooleanAttributeValue struct {
	Attribute
	Value bool
}

func (a BooleanAttributeValue) InspectValue() string {
	return fmt.Sprintf("%t", a.Value)
}

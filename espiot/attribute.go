package espiot

import (
	"fmt"
	"strconv"
	"time"
)

type AttributeAndValue interface {
	AttributeDef() Attribute
	InspectValue() string
	Interface() interface{}
	accept(string) error
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

func (a *StringAttributeValue) accept(s string) error {
	a.Attribute.UpdatedAt = time.Now()
	a.Value = s
	return nil
}

func (a StringAttributeValue) InspectValue() string {
	return a.Value
}

func (a StringAttributeValue) Interface() interface{} {
	return a.Value
}

type BooleanAttributeValue struct {
	Attribute
	Value bool
}

func (a BooleanAttributeValue) InspectValue() string {
	return fmt.Sprintf("%t", a.Value)
}

func (a *BooleanAttributeValue) accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.ParseBool(s)
	return
}

func (a BooleanAttributeValue) Interface() interface{} {
	return a.Value
}

type IntegerAttributeValue struct {
	Attribute
	Value int
}

func (a *IntegerAttributeValue) accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.Atoi(s)
	return
}

func (a IntegerAttributeValue) InspectValue() string {
	return fmt.Sprintf("%d", a.Value)
}

func (a IntegerAttributeValue) Interface() interface{} {
	return a.Value
}

type DoubleAttributeValue struct {
	Attribute
	Value float64
}

func (a *DoubleAttributeValue) accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.ParseFloat(s, 64)
	return
}

func (a DoubleAttributeValue) InspectValue() string {
	return fmt.Sprintf("%f", a.Value)
}

func (a DoubleAttributeValue) Interface() interface{} {
	return a.Value
}

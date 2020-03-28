package espiot

import (
	"fmt"
	"strconv"
	"time"
)

type AttributeAndValue interface {
	AttributeDef() Attribute
	InspectValue() string
	Accept(string) error
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

func (a *StringAttributeValue) Accept(s string) error {
	a.Attribute.UpdatedAt = time.Now()
	a.Value = s
	return nil
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

func (a *BooleanAttributeValue) Accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.ParseBool(s)
	return
}

type IntegerAttributeValue struct {
	Attribute
	Value int
}

func (a *IntegerAttributeValue) Accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.Atoi(s)
	return
}

func (a IntegerAttributeValue) InspectValue() string {
	return fmt.Sprintf("%d", a.Value)
}

type DoubleAttributeValue struct {
	Attribute
	Value float64
}

func (a *DoubleAttributeValue) Accept(s string) (err error) {
	a.Attribute.UpdatedAt = time.Now()
	a.Value, err = strconv.ParseFloat(s, 64)
	return
}

func (a DoubleAttributeValue) InspectValue() string {
	return fmt.Sprintf("%f", a.Value)
}

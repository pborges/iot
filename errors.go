package iot

import (
	"errors"
	"fmt"
)

var ErrDuplicateName = errors.New("duplicate name")

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e)
}

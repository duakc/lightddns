package adapter

import (
	"errors"
	"fmt"
)

var ErrRequireToken = errors.New("token required")

type ManagedNotFoundError struct {
	Name string
}

func (e *ManagedNotFoundError) Error() string {
	return fmt.Sprintf("`%s` not found", e.Name)
}

type EmptyGroupError struct {
	Type string
	Name string
}

func (e *EmptyGroupError) Error() string {
	return fmt.Sprintf("empty %s group: %s", e.Type, e.Name)
}

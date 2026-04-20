package providers

import (
	"errors"
	"fmt"
)

var (
	ErrRequireToken = errors.New("token required")
)

type ProviderNotFoundError struct {
	Name string
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("provider `%s` not found", e.Name)
}

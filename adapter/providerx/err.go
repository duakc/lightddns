package providerx

import "fmt"

type ProviderNotFoundError struct {
	Err error
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("provider not found: %s", e.Err.Error())
}

func (e *ProviderNotFoundError) Unwrap() error {
	return e.Err
}

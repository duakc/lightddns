package adapter

import (
	"fmt"

	"github.com/duakc/mt/services"
)

type Service interface {
	managedType
	services.LifeCycle
}

type (
	ServiceManager = Manager[Service]
)

var ServiceRegistry = NewRegister[Service]()

type ServiceNotFoundError struct {
	*ManagedNotFoundError
}

func (e *ServiceNotFoundError) Error() string {
	return fmt.Sprintf("service: %s", e.ManagedNotFoundError.Error())
}

package adapter

import (
	"fmt"

	"github.com/duakc/mt/services"
	"go.uber.org/zap"
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

func CreateServiceLogger(logger *zap.Logger, srv Service) *zap.Logger {
	return logger.With(
		zap.String("service", srv.Name()),
		zap.String("service_type", srv.Type())).
		Named("service")
}

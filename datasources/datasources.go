package datasources

import (
	"fmt"

	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func NewLogger(logger *zap.Logger, option options.AbstractDatasourceOption) *zap.Logger {
	return logger.With(
		zap.String("type", "datasource"),
		zap.String("datasource_type", option.Type)).
		Named(option.Name)
}

type DatasourceNotFoundError struct {
	Name string
}

func (e *DatasourceNotFoundError) Error() string {
	return fmt.Sprintf("datasource `%s` not found", e.Name)
}

type EmptyGroupError struct {
	Type string
	Name string
}

func (e *EmptyGroupError) Error() string {
	return fmt.Sprintf("empty %s group: %s", e.Type, e.Name)
}

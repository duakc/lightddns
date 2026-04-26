package datasources

import (
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func NewLogger(logger *zap.Logger, option options.AbstractDatasourceOption) *zap.Logger {
	return logger.With(
		zap.String("type", "datasource"),
		zap.String("datasource_type", option.Type)).
		Named(option.Name)
}

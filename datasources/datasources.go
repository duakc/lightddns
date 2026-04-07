package datasources

import (
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
)

func NewLogger(logger *zap.Logger, option options.AbstractDatasourceOption) *zap.Logger {
	return zaplog.ExtendName(logger, option.Name).With(
		zap.String("type", "datasource"),
		zap.String("datasource_type", option.Type))
}

package lightddns

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LightDDNS struct {
	logger *zap.Logger

	providerManager   *adapter.ProviderManager
	datasourceManager *adapter.DataSourceManager
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	logger, err := newLoggerWithOptions(opt.Log)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}
	ctx = ctxservice.NewRegistry(ctx, ctxservice.NewDefaultRegistry())
	ctxservice.Store(ctx, zaplog.LoggerKey{}, logger)

	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	dataSourceManager := adapter.NewManager[adapter.DataSource](adapter.DataSourceRegister)

	for i := 0; i < len(opt.Providers); i++ {
		providerOption := opt.Providers[i]
		if err := providerManager.Create(ctx, providerOption.Type, providerOption.Option); err != nil {
			return nil, fmt.Errorf("create provider: %w", err)
		}
	}
	for i := 0; i < len(opt.DataSources); i++ {
		datasourceOption := opt.DataSources[i]
		if err := dataSourceManager.Create(ctx, datasourceOption.Type, datasourceOption.Option); err != nil {
			return nil, fmt.Errorf("create datasource: %w", err)
		}
	}
	ddns := &LightDDNS{
		logger:            logger,
		providerManager:   (*adapter.ProviderManager)(providerManager),
		datasourceManager: (*adapter.DataSourceManager)(dataSourceManager),
	}

	return ddns, nil
}

func newLoggerWithOptions(opt options.OptionLog) (*zap.Logger, error) {
	if opt.Disabled {
		return zaplog.NOP, nil
	}

	level, err := zapcore.ParseLevel(opt.Level)
	if err != nil {
		return nil, err
	}
	var outputFD *os.File
	switch strings.ToLower(opt.Output) {
	case "stdout", "":
		outputFD = os.Stdout
	case "stderr":
		outputFD = os.Stderr
	default:
		outputFD, err = os.OpenFile(opt.Output, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}
	}
	logger := zaplog.NewDefault(outputFD, level, nil)
	return logger, nil
}

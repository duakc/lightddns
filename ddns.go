package lightddns

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LightDDNS struct {
	logger  *zap.Logger
	domains []*Domain

	providerManager   *adapter.ProviderManager
	datasourceManager *adapter.DatasourceManager
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	logger, err := newLoggerWithOptions(opt.Log)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	defer logger.Sync()

	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	dataSourceManager := adapter.NewManager[adapter.Datasource](adapter.DatasourceRegister)
	lookctx.Store[zaplog.LoggerKey](ctx, logger)
	lookctx.Store[adapter.ProviderManagerKey](ctx, providerManager)
	lookctx.Store[adapter.DatasourceManagerKey](ctx, dataSourceManager)

	logger = zaplog.ExtendName(logger, "main")

	for i := 0; i < len(opt.Providers); i++ {
		providerOption := opt.Providers[i]
		if err := providerManager.Create(ctx, providerOption.Type, providerOption.Option); err != nil {
			return nil, fmt.Errorf("create provider[%d]: %w", i, err)
		}
		logger.Debug("new provider created", zap.String("type", providerOption.Type),
			zap.String("name", providerOption.Name))
	}
	for i := 0; i < len(opt.DataSources); i++ {
		datasourceOption := opt.DataSources[i]
		if err := dataSourceManager.Create(ctx, datasourceOption.Type, datasourceOption.Option); err != nil {
			return nil, fmt.Errorf("create datasource[%d]: %w", i, err)
		}
		logger.Debug("new datasource created", zap.String("type", datasourceOption.Type),
			zap.String("name", datasourceOption.Name))
	}
	var domains []*Domain
	for i := 0; i < len(opt.Domains); i++ {
		domain, err := NewDomain(ctx, opt.Domains[i])
		if err != nil {
			return nil, fmt.Errorf("create domain[%d]: %w", i, err)
		}
		domains = append(domains, domain)
	}

	ddns := &LightDDNS{
		providerManager:   providerManager,
		datasourceManager: dataSourceManager,
		logger:            logger,
		domains:           domains,
	}

	return ddns, nil
}

func (ddns *LightDDNS) Once(ctx context.Context) {
	logger := ddns.logger
	defer logger.Sync()

	for i := 0; i < len(ddns.domains); i++ {
		domain := ddns.domains[i]
		err := domain.UpdateOnce(ctx)
		if err != nil {
			logger.Error("update failed", zap.Error(err))
			return
		}
	}
}

func (ddns *LightDDNS) Start(ctx context.Context) error {
	logger := ddns.logger
	defer logger.Sync()

	for i := 0; i < len(ddns.domains); i++ {
		domain := ddns.domains[i]
		go domain.UpdateLoop(ctx)
	}
	logger.Info("started")
	<-ctx.Done()
	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func newLoggerWithOptions(opt options.LogOption) (*zap.Logger, error) {
	if opt.Disabled {
		return zaplog.NOP, nil
	}
	var (
		level = zapcore.InfoLevel
		err   error
	)
	if len(opt.Level) != 0 {
		level, err = zapcore.ParseLevel(opt.Level)
		if err != nil {
			return nil, err
		}
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

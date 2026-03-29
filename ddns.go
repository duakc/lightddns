package lightddns

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LightDDNS struct {
	logger  *zap.Logger
	domains []*Domain

	providerManager   *adapter.ProviderManager
	datasourceManager *adapter.DataSourceManager
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	logger, err := newLoggerWithOptions(opt.Log)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}
	ctx = ctxservice.NewRegistry(ctx, ctxservice.NewDefaultRegistry())

	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	dataSourceManager := adapter.NewManager[adapter.DataSource](adapter.DataSourceRegister)
	ctxservice.Store(ctx, common.Zero[zaplog.LoggerKey](), logger)
	ctxservice.Store(ctx, common.Zero[adapter.ProviderManagerKey](), providerManager)
	ctxservice.Store(ctx, common.Zero[adapter.DataSourceManagerKey](), dataSourceManager)

	for i := 0; i < len(opt.Providers); i++ {
		providerOption := opt.Providers[i]
		if err := providerManager.Create(ctx, providerOption.Type, providerOption.Option); err != nil {
			return nil, fmt.Errorf("create provider[%d]: %w", i, err)
		}
	}
	for i := 0; i < len(opt.DataSources); i++ {
		datasourceOption := opt.DataSources[i]
		if err := dataSourceManager.Create(ctx, datasourceOption.Type, datasourceOption.Option); err != nil {
			return nil, fmt.Errorf("create datasource[%d]: %w", i, err)
		}
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
		logger:  zaplog.ExtendName(logger, "main"),
		domains: domains,

		providerManager:   (*adapter.ProviderManager)(providerManager),
		datasourceManager: (*adapter.DataSourceManager)(dataSourceManager),
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
	logger.Warn("started")
	<-ctx.Done()
	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func newLoggerWithOptions(opt options.OptionLog) (*zap.Logger, error) {
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

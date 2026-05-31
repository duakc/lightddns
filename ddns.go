package lightddns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/closeme"
	"github.com/duakc/mt/services/container"
	"github.com/duakc/mt/services/filehelper"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Root-level metric names (no subsystem). Final name composed with
// prometheus.BuildFQName(ddnsmetric.Namespace, "", <leaf>).
const (
	metricBuildInfo = "build_info"
)

type LightDDNS struct {
	logger  *zap.Logger
	domains []*Domain

	datasources []adapter.Datasource
	providers   []adapter.Provider
	services    []adapter.Service

	providerManager   adapter.ProviderManager
	datasourceManager adapter.DatasourceManager
	serviceManager    adapter.ServiceManager
	metricsRegistry   metrics.Registry
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	logger, err := newLoggerWithOptions(ctx, opt.Log)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	metricsRegistry := metrics.New(prometheusEnabled(opt.Services))
	services.Store[metrics.Registry](ctx, metricsRegistry)

	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	services.Store[adapter.ProviderManager](ctx, providerManager)
	datasourceManager := adapter.NewManager[adapter.Datasource](adapter.DatasourceRegister)
	services.Store[adapter.DatasourceManager](ctx, datasourceManager)
	serviceManager := adapter.NewManager[adapter.Service](adapter.ServiceRegistry)
	services.Store[adapter.ServiceManager](ctx, serviceManager)

	initLogger := logger.Named("init")
	resortedDatasources, err := resortDatasources(opt.Datasources)
	if err != nil {
		return nil, fmt.Errorf("resort: %w", err)
	}

	var (
		datasources []adapter.Datasource
		providers   []adapter.Provider
		services    []adapter.Service
	)

	for i := 0; i < len(resortedDatasources); i++ {
		var (
			datasource adapter.Datasource
			err        error
		)

		datasourceOption := resortedDatasources[i]
		if datasource, err = datasourceManager.Create(ctx, createDatasourceLogger(logger, datasourceOption),
			datasourceOption.Type, datasourceOption.Option); err != nil {
			return nil, fmt.Errorf("create datasource `%s,type=%s` failed: %w",
				datasourceOption.Name, datasourceOption.Type, err)
		}

		datasources = append(datasources, datasource)
		initLogger.Info("new datasource created",
			zap.String("type", datasourceOption.Type),
			zap.String("name", datasourceOption.Name))
	}

	for i := 0; i < len(opt.Providers); i++ {
		var (
			provider adapter.Provider
			err      error
		)
		providerOption := opt.Providers[i]
		if provider, err = providerManager.Create(ctx, creatProviderLogger(logger, providerOption),
			providerOption.Type, providerOption.Option); err != nil {
			return nil, fmt.Errorf("create provider `%s,type=%s` failed: %w",
				providerOption.Name, providerOption.Type, err)
		}
		providers = append(providers, provider)
		initLogger.Info("new provider created",
			zap.String("type", providerOption.Type),
			zap.String("name", providerOption.Name))
	}

	for i := 0; i < len(opt.Services); i++ {
		var (
			service adapter.Service
			err     error
		)

		serviceOption := opt.Services[i]
		if service, err = serviceManager.Create(ctx, createServiceLogger(logger, serviceOption),
			serviceOption.Type, serviceOption.Option); err != nil {
			if errors.Is(err, adapter.ErrManagedItemNotEnabled) {
				initLogger.Warn("service not enabled",
					zap.String("type", serviceOption.Type),
					zap.String("name", serviceOption.Name))
				continue
			}
			return nil, fmt.Errorf("create service `%s,type=%s` failed: %w",
				serviceOption.Name, serviceOption.Type, err)
		}
		services = append(services, service)
		initLogger.Info("new service created",
			zap.String("type", serviceOption.Type),
			zap.String("name", serviceOption.Name))
	}

	var domains []*Domain

	for i := 0; i < len(opt.Domains); i++ {
		domainOption := opt.Domains[i]
		domain, err := NewDomain(ctx,
			logger.With(zap.String("domain", string(domainOption.Domain))), domainOption)
		if domain == nil && err == nil {
			// not enabled
			initLogger.Warn("domain not enabled", zap.String("domain", string(domainOption.Domain)))
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("create domain `%s`: %w", domainOption.Domain, err)
		}
		domains = append(domains, domain)
	}

	ld := &LightDDNS{
		logger:            logger,
		providerManager:   providerManager,
		datasourceManager: datasourceManager,
		serviceManager:    serviceManager,
		metricsRegistry:   metricsRegistry,

		domains:     domains,
		datasources: datasources,
		providers:   providers,
		services:    services,
	}

	return ld, nil
}

func (ld *LightDDNS) StartOnce(ctx context.Context, fastfail bool) error {
	var err error
	for i := 0; i < len(ld.domains); i++ {
		domain := ld.domains[i]
		updateErr := domain.Update(ctx)
		if fastfail && updateErr != nil {
			return updateErr
		}
		err = errors.Join(err, updateErr)
	}
	return err
}

func (ld *LightDDNS) Start(ctx context.Context, stage services.Stage) error {
	if stage == services.StagePreStart {
		var cont container.Container
		containerProvider := services.Lookup[container.Provider](ctx)
		ctx, cont = containerProvider.New(ctx)
		defer containerProvider.Release(cont)
		cont.IncRef()
		defer cont.DecRef()

		ddnsmetric.Pass(ctx, ld.metricsRegistry)

		setupPrometheus(ld.metricsRegistry)
	}
	var err error
	for i := 0; i < len(ld.services); i++ {
		err = errors.Join(err, services.Start(ctx, stage, ld.services[i]))
	}
	for i := 0; i < len(ld.domains); i++ {
		domain := ld.domains[i]
		err = errors.Join(err, domain.Start(ctx, stage))
	}
	return err
}

// setupPrometheus registers root-level metrics (those without a per-variant
// subsystem). Called once from Start(PreStart) — kept out of New() so the
// check command's bare-New path doesn't touch prometheus state.
func setupPrometheus(reg metrics.Registry) {
	reg.GaugeVec(
		prometheus.BuildFQName(ddnsmetric.Namespace, "", metricBuildInfo),
		"Build information. Value is always 1.",
		[]string{constpkg.MetricLabelVersion, constpkg.MetricLabelBranch},
	).With(constpkg.Version, constpkg.Branch).Set(1)
}

func (ld *LightDDNS) Close() error {
	var err error
	for i := 0; i < len(ld.domains); i++ {
		domain := ld.domains[i]
		err = errors.Join(err, domain.Close())
	}
	for i := len(ld.services) - 1; i >= 0; i-- {
		err = errors.Join(err, services.CloseService(ld.services[i]))
	}
	return err
}

// resortDatasources returns datasources in initialization order using Kahn's
// topological sort (O(V+E)). If A.Group() returns ["B"], B is initialized
// before A. Relative order among independent nodes follows the original input.
// Steps:
//  1. Build name→option index and per-node dependency list.
//  2. Validate each dep exists; compute in-degree and reverse-adjacency table.
//  3. Seed queue with in-degree-0 nodes (no deps), preserving input order.
//  4. BFS: dequeue → append to result → decrement in-degree of dependents
//     via rdeps → enqueue newly-zero nodes.
//  5. len(result) < len(ds) implies a cycle.
func resortDatasources(ds []options.DatasourceOption) ([]options.DatasourceOption, error) {
	dsMap := make(map[string]options.DatasourceOption, len(ds))
	for _, d := range ds {
		dsMap[d.Name] = d
	}

	deps := make(map[string][]string, len(ds))
	for _, d := range ds {
		if g, ok := d.Option.(options.DatasourceGrouper); ok {
			deps[d.Name] = g.Group()
		}
	}

	inDegree := make(map[string]int, len(ds))
	for _, d := range ds {
		// ensure every node is present
		inDegree[d.Name] = 0
	}

	// rdeps[x] = nodes that directly depend on x (built in input order)
	rdeps := make(map[string][]string, len(ds))
	for _, d := range ds {
		for _, dep := range deps[d.Name] {
			if _, exists := dsMap[dep]; !exists {
				return nil, fmt.Errorf("datasource(%s) depends on unknown datasource: %q", d.Name, dep)
			}
			inDegree[d.Name]++
			rdeps[dep] = append(rdeps[dep], d.Name)
		}
	}

	// seed queue (input order preserved)
	queue := make([]string, 0, len(ds))
	for _, d := range ds {
		if inDegree[d.Name] == 0 {
			queue = append(queue, d.Name)
		}
	}

	// BFS
	result := make([]options.DatasourceOption, 0, len(ds))
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, dsMap[cur])

		// O(E) total across all iterations — each edge visited exactly once
		for _, dependent := range rdeps[cur] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// cycle check
	if len(result) != len(ds) {
		return nil, fmt.Errorf("circular dependency detected among datasources")
	}

	return result, nil
}

func newLoggerWithOptions(ctx context.Context, opt options.LogOption) (*zap.Logger, error) {
	if opt.Disabled {
		return zaplog.NOP, nil
	}
	var (
		fileHelper = services.Lookup[filehelper.Helper](ctx)
		level      = zapcore.Level(opt.Level)
		err        error
	)
	zaplog.DefaultLevel(level)
	closeManager := services.Lookup[closeme.Manager](ctx)

	var outputFD io.Writer
	switch strings.ToLower(opt.Output) {
	case "stdout", "":
		outputFD = zaplog.Stdout
	case "stderr":
		outputFD = zaplog.Stderr
	default:
		outputFD, err = fileHelper.Create(opt.Output)
		if err != nil {
			return nil, err
		}
	}

	logger := zaplog.NewDefault(closeManager, outputFD, level, nil)
	return logger, nil
}

func prometheusEnabled(svcs []options.ServiceOption) bool {
	for i := 0; i < len(svcs); i++ {
		if svcs[i].Type != constpkg.ServiceTypePrometheus {
			continue
		}
		if po, ok := svcs[i].Option.(*options.PrometheusServiceOption); ok && po.Enabled {
			return true
		}
	}
	return false
}

func createServiceLogger(logger *zap.Logger, srv options.ServiceOption) *zap.Logger {
	return logger.With(
		zap.String("service", srv.Name),
		zap.String("service_type", srv.Type)).
		Named("service")
}

func creatProviderLogger(logger *zap.Logger, provider options.ProviderOption) *zap.Logger {
	return logger.With(
		zap.String("provider", provider.Name),
		zap.String("provider_type", provider.Type)).
		Named("provider")
}

func createDatasourceLogger(logger *zap.Logger, datasource options.DatasourceOption) *zap.Logger {
	return logger.With(
		zap.String("datasource", datasource.Name),
		zap.String("datasource_type", datasource.Type)).
		Named("datasource")
}

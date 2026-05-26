package lightddns

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/closeme"
	"github.com/duakc/mt/services/container"
	"github.com/duakc/mt/services/filehelper"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LightDDNS struct {
	logger  *zap.Logger
	domains []*Domain

	providerManager   adapter.ProviderManager
	datasourceManager adapter.DatasourceManager
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	logger, err := newLoggerWithOptions(ctx, opt.Log)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	loggerFactory := zaplog.NewFactory(logger)
	services.Store[zaplog.Factory](ctx, loggerFactory)
	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	services.Store[adapter.ProviderManager](ctx, providerManager)
	datasourceManager := adapter.NewManager[adapter.Datasource](adapter.DatasourceRegister)
	services.Store[adapter.DatasourceManager](ctx, datasourceManager)

	// pass logger to downstream
	containerProvider := services.Lookup[container.Provider](ctx)
	ctx = containerProvider.New(ctx)
	defer containerProvider.Release(ctx)
	zaplog.WithContext(ctx, logger)

	logger = logger.Named("main")
	resortedDatasources, err := resortDatasources(opt.Datasources)
	if err != nil {
		return nil, fmt.Errorf("resort: %w", err)
	}

	for i := 0; i < len(resortedDatasources); i++ {
		datasourceOption := resortedDatasources[i]
		if err := datasourceManager.Create(ctx, datasourceOption.Type, datasourceOption.Option); err != nil {
			return nil, fmt.Errorf("create datasource `%s,type=%s` failed: %w",
				datasourceOption.Name, datasourceOption.Type, err)
		}
		logger.Debug("new datasource created", zap.String("type", datasourceOption.Type),
			zap.String("name", datasourceOption.Name))
	}
	for i := 0; i < len(opt.Providers); i++ {
		providerOption := opt.Providers[i]
		if err := providerManager.Create(ctx, providerOption.Type, providerOption.Option); err != nil {
			return nil, fmt.Errorf("create provider `%s,type=%s` failed: %w",
				providerOption.Name, providerOption.Type, err)
		}
		logger.Debug("new provider created", zap.String("type", providerOption.Type),
			zap.String("name", providerOption.Name))
	}

	var (
		domains               []*Domain
		defaultProviderName   string
		defaultDatasourceName string
	)

	if len(opt.Providers) == 1 {
		defaultProviderName = opt.Providers[0].Name
	}
	if len(opt.Datasources) == 1 {
		defaultDatasourceName = opt.Datasources[0].Name
	}

	for i := 0; i < len(opt.Domains); i++ {
		domainOption := opt.Domains[i]
		domainOption.Provider = cmp.Or(domainOption.Provider, defaultProviderName)
		domainOption.Datasource = cmp.Or(domainOption.Datasource, defaultDatasourceName)
		domain, err := NewDomain(ctx, domainOption)
		if domain == nil && err == nil {
			// not enabled
			logger.Warn("domain not enabled", zap.String("domain", string(domainOption.Domain)))
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("create domain `%s`: %w", domainOption.Domain, err)
		}
		domains = append(domains, domain)
	}

	ld := &LightDDNS{
		logger:            logger,
		domains:           domains,
		providerManager:   providerManager,
		datasourceManager: datasourceManager,
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
	var err error
	for i := 0; i < len(ld.domains); i++ {
		domain := ld.domains[i]
		err = errors.Join(err, domain.Start(ctx, stage))
	}
	return err
}

func (ld *LightDDNS) Close() error {
	var err error
	for i := 0; i < len(ld.domains); i++ {
		domain := ld.domains[i]
		err = errors.Join(err, domain.Close())
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

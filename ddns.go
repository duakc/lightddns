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

	providerManager := adapter.NewManager[adapter.Provider](adapter.ProviderRegister)
	dataSourceManager := adapter.NewManager[adapter.Datasource](adapter.DatasourceRegister)
	lookctx.Store[zaplog.LoggerKey](ctx, logger)
	lookctx.Store[adapter.ProviderManagerKey](ctx, providerManager)
	lookctx.Store[adapter.DatasourceManagerKey](ctx, dataSourceManager)

	logger = logger.Named("main")
	resortedDatasources, err := resortDatasources(opt.DataSources)
	if err != nil {
		return nil, fmt.Errorf("resort: %w", err)
	}

	for i := 0; i < len(resortedDatasources); i++ {
		datasourceOption := resortedDatasources[i]
		if err := dataSourceManager.Create(ctx, datasourceOption.Type, datasourceOption.Option); err != nil {
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
	var domains []*Domain
	for i := 0; i < len(opt.Domains); i++ {
		domain, err := NewDomain(ctx, opt.Domains[i])
		if err != nil && errors.Is(err, errDomainNotEnabled) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("create domain[%d]: %w", i, err)
		}
		domains = append(domains, domain)
	}

	ddns := &LightDDNS{
		logger:            logger,
		domains:           domains,
		providerManager:   providerManager,
		datasourceManager: dataSourceManager,
	}

	return ddns, nil
}

func (ddns *LightDDNS) Once(ctx context.Context) {
	logger := ddns.logger

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
		outputFD, err = os.OpenFile(opt.Output, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
		if err != nil {
			return nil, err
		}
	}
	logger := zaplog.NewDefault(outputFD, level, nil)
	return logger, nil
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
		if g, ok := d.Option.(options.DatasourceIsGroup); ok {
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

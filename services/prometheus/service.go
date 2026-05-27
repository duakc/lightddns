package prometheus

import (
	"context"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"
)

func init() {
	adapter.Register(
		adapter.ServiceRegistry,
		constpkg.ServiceTypePrometheus,
		New,
	)
}

type Prometheus struct{}

func New(ctx context.Context, option options.PrometheusServiceOption) (adapter.Service, error) {
	panic("implement me")
}

func (s *Prometheus) Start(ctx context.Context, stage services.Stage) error {
	panic("implement me")
}

func (s *Prometheus) Close() error {
	panic("implement me")
}

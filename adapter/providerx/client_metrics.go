package providerx

import (
	"context"
	"time"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/metricx"

	"github.com/duakc/mt/services"
)

type operationMetrics struct {
	total    interface{ Inc() }
	failure  interface{ Inc() }
	duration interface{ Observe(float64) }
}

var _ ddnsx.DDNSClient[int] = (*MetricsClient[int])(nil)

type MetricsClient[R any] struct {
	next ddnsx.DDNSClient[R]

	resolveZone operationMetrics
	listRecords operationMetrics
	create      operationMetrics
	update      operationMetrics
	delete      operationMetrics
}

func NewMetricsClientFromContext[R any](
	ctx context.Context,
	clientName, clientType string,
	next ddnsx.DDNSClient[R],
) *MetricsClient[R] {
	return NewMetricsClient(
		services.Lookup[metricx.ProviderFactory](ctx), clientName, clientType, next,
	)
}

func NewMetricsClient[R any](
	factory metricx.ProviderFactory,
	clientName, clientType string,
	next ddnsx.DDNSClient[R],
) *MetricsClient[R] {
	operation := func(name string) operationMetrics {
		return operationMetrics{
			total:    factory.OperationTotal(clientName, clientType, name),
			failure:  factory.OperationFailure(clientName, clientType, name),
			duration: factory.OperationDuration(clientName, clientType, name, nil),
		}
	}
	return &MetricsClient[R]{
		next:        next,
		resolveZone: operation(OperationResolveZone),
		listRecords: operation(OperationListRecords),
		create:      operation(OperationCreateRecord),
		update:      operation(OperationUpdateRecord),
		delete:      operation(OperationDeleteRecord),
	}
}

func (c *MetricsClient[R]) ResolveZone(ctx context.Context, fqdn string) (zone ddnsx.Zone, err error) {
	done := c.record(c.resolveZone)
	defer done(&err)
	return c.next.ResolveZone(ctx, fqdn)
}

func (c *MetricsClient[R]) Records(ctx context.Context, key ddnsx.RecordKey) (records []ddnsx.Existing[R], err error) {
	done := c.record(c.listRecords)
	defer done(&err)
	return c.next.Records(ctx, key)
}

func (c *MetricsClient[R]) Create(ctx context.Context, target ddnsx.RecordSpec) (err error) {
	done := c.record(c.create)
	defer done(&err)
	return c.next.Create(ctx, target)
}

func (c *MetricsClient[R]) Update(ctx context.Context, target ddnsx.RecordSpec, record R) (err error) {
	done := c.record(c.update)
	defer done(&err)
	return c.next.Update(ctx, target, record)
}

func (c *MetricsClient[R]) Delete(ctx context.Context, key ddnsx.RecordKey, record R) (err error) {
	done := c.record(c.delete)
	defer done(&err)
	return c.next.Delete(ctx, key, record)
}

func (*MetricsClient[R]) record(metrics operationMetrics) func(*error) {
	started := time.Now()
	return func(errp *error) {
		metrics.total.Inc()
		metrics.duration.Observe(time.Since(started).Seconds())
		if errp != nil && *errp != nil {
			metrics.failure.Inc()
		}
	}
}

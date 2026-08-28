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

type MetricsClient[T ddnsx.DDNSRecordComparable[T]] struct {
	next ddnsx.DDNSClient[T]

	resolveZone operationMetrics
	listRecords operationMetrics
	create      operationMetrics
	update      operationMetrics
	delete      operationMetrics
}

func NewMetricsClientFromContext[T ddnsx.DDNSRecordComparable[T]](
	ctx context.Context,
	clientName, clientType string,
	next ddnsx.DDNSClient[T],
) *MetricsClient[T] {
	return NewMetricsClient(
		services.Lookup[metricx.ProviderFactory](ctx), clientName, clientType, next,
	)
}

func NewMetricsClient[T ddnsx.DDNSRecordComparable[T]](
	factory metricx.ProviderFactory,
	clientName, clientType string,
	next ddnsx.DDNSClient[T],
) *MetricsClient[T] {
	operation := func(name string) operationMetrics {
		return operationMetrics{
			total:    factory.OperationTotal(clientName, clientType, name),
			failure:  factory.OperationFailure(clientName, clientType, name),
			duration: factory.OperationDuration(clientName, clientType, name, nil),
		}
	}
	return &MetricsClient[T]{
		next:        next,
		resolveZone: operation(OperationResolveZone),
		listRecords: operation(OperationListRecords),
		create:      operation(OperationCreateRecord),
		update:      operation(OperationUpdateRecord),
		delete:      operation(OperationDeleteRecord),
	}
}

func (c *MetricsClient[T]) ResolveZone(ctx context.Context, fqdn string) (zone ddnsx.Zone, err error) {
	done := c.record(c.resolveZone)
	defer done(&err)
	return c.next.ResolveZone(ctx, fqdn)
}

func (c *MetricsClient[T]) Records(ctx context.Context, key ddnsx.RecordKey) (records []T, err error) {
	done := c.record(c.listRecords)
	defer done(&err)
	return c.next.Records(ctx, key)
}

func (c *MetricsClient[T]) Create(ctx context.Context, target ddnsx.RecordSpec, desired T) (err error) {
	done := c.record(c.create)
	defer done(&err)
	return c.next.Create(ctx, target, desired)
}

func (c *MetricsClient[T]) Update(ctx context.Context, target ddnsx.RecordSpec, desired T, existing T) (err error) {
	done := c.record(c.update)
	defer done(&err)
	return c.next.Update(ctx, target, desired, existing)
}

func (c *MetricsClient[T]) Delete(ctx context.Context, key ddnsx.RecordKey, record T) (err error) {
	done := c.record(c.delete)
	defer done(&err)
	return c.next.Delete(ctx, key, record)
}

func (*MetricsClient[T]) record(metrics operationMetrics) func(*error) {
	started := time.Now()
	return func(errp *error) {
		metrics.total.Inc()
		metrics.duration.Observe(time.Since(started).Seconds())
		if errp != nil && *errp != nil {
			metrics.failure.Inc()
		}
	}
}

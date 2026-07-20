package ddnsx

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/infra/zaplog"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

type Reconciler[R any] struct {
	logger *zap.Logger
	client DDNSClient[R]
}

func NewReconciler[R any](logger *zap.Logger, client DDNSClient[R]) *Reconciler[R] {
	return &Reconciler[R]{
		logger: zaplog.DoNotPanic(logger).Named("reconciler"),
		client: client,
	}
}

func (r *Reconciler[R]) Diff(ctx context.Context, fqdn string, addr []netip.Addr) (bool, error) {
	diffs, err := r.diffs(ctx, fqdn, addr)
	if err != nil {
		return false, err
	}
	return len(diffs) > 0, nil
}

func (r *Reconciler[R]) Update(ctx context.Context, fqdn string, ttl uint32, addr []netip.Addr) (bool, error) {
	logger := r.logger.With(zap.String("domain", fqdn))
	logger.Debug("new update request",
		zap.Stringers("addresses", addr),
	)

	key, err := r.resolve(ctx, fqdn)
	if err != nil {
		return false, err
	}

	diffs, err := BuildDiffs(ctx, key, addr, r.client)
	if err != nil {
		return false, fmt.Errorf("diff: %w", err)
	}
	if len(diffs) == 0 {
		logger.Info("no difference since last updated, skip")
		return false, nil
	}

	changed := false
	for _, diff := range diffs {
		if err := r.applyDiff(ctx, logger, key, ttl, diff); err != nil {
			return changed, err
		}
		changed = true
	}
	return changed, nil
}

func (r *Reconciler[R]) diffs(ctx context.Context, fqdn string, addr []netip.Addr) ([]Diff[R], error) {
	key, err := r.resolve(ctx, fqdn)
	if err != nil {
		return nil, err
	}
	return BuildDiffs(ctx, key, addr, r.client)
}

func (r *Reconciler[R]) resolve(ctx context.Context, fqdn string) (RecordKey, error) {
	fqdn = mDns.Fqdn(fqdn)

	zone, err := r.client.ResolveZone(ctx, fqdn)
	if err != nil {
		return RecordKey{}, fmt.Errorf("resolve zone for %s: %w", fqdn, err)
	}
	if !zone.Valid() {
		return RecordKey{}, fmt.Errorf("resolve zone for %s: %w", fqdn, ErrZoneNotFound)
	}
	return RecordKey{FQDN: fqdn, Zone: zone}, nil
}

func (r *Reconciler[R]) applyDiff(ctx context.Context, logger *zap.Logger,
	key RecordKey, ttl uint32, diff Diff[R],
) error {
	fields := []zap.Field{
		zap.String("domain", diff.Domain),
		zap.Stringer("action", diff.Action),
	}
	if diff.Target.IsValid() {
		fields = append(fields, zap.Stringer("target", diff.Target))
	}
	if diff.Source.IsValid() {
		fields = append(fields, zap.Stringer("source", diff.Source))
	}
	logger = logger.WithLazy(fields...)
	key.Type = diff.Type

	var err error
	switch diff.Action {
	case DDNSActionCreate:
		logger.Info("create")
		err = r.client.Create(ctx, RecordSpec{
			RecordKey: key,
			Address:   diff.Target.Unmap(),
			TTL:       ttl,
		})
	case DDNSActionUpdate:
		logger.Info("update")
		err = r.client.Update(ctx, RecordSpec{
			RecordKey: key,
			Address:   diff.Target.Unmap(),
			TTL:       ttl,
		}, diff.Record)
	case DDNSActionDelete:
		logger.Info("delete")
		err = r.client.Delete(ctx, key, diff.Record)
	default:
		return fmt.Errorf("unknown DDNS action: %s", diff.Action)
	}
	if err != nil {
		return fmt.Errorf("%s %s record for %s: %w", diff.Action, diff.Type, diff.Domain, err)
	}
	return nil
}

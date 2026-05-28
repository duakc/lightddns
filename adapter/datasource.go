package adapter

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"slices"

	"github.com/duakc/lightddns/infra/netool"
	"go.uber.org/zap"
)

type Datasource interface {
	managedType
	IP(context.Context) ([]netip.Addr, error)
}

type DatasourceDualStack interface {
	Datasource
	IPv4(ctx context.Context) ([]netip.Addr, error)
	IPv6(ctx context.Context) ([]netip.Addr, error)
}

type (
	DatasourceManager = Manager[Datasource]
)

var DatasourceRegister = NewRegister[Datasource]()

func MergeDualStackDatasourceIP(ctx context.Context, s DatasourceDualStack) ([]netip.Addr, error) {
	ipv4, err := s.IPv4(ctx)
	if err != nil {
		return nil, fmt.Errorf("ipv4: %w", err)
	}

	ipv6, err := s.IPv6(ctx)
	if err != nil {
		return nil, fmt.Errorf("ipv6: %w", err)
	}
	return append(ipv4, ipv6...), nil
}

func MergeDatasources(ctx context.Context, datasources []Datasource, ipv4, ipv6, fastfail bool) ([]netip.Addr, error) {
	if !ipv4 && !ipv6 {
		// nothing to do
		return []netip.Addr{}, nil
	}
	var outerErr error
	merged := make(map[netip.Addr]struct{})

	addIPs := func(ds Datasource, version string, getIPs func(context.Context) ([]netip.Addr, error)) error {
		addr, err := getIPs(ctx)
		if err != nil {
			err = newDatasourceError(err, version, ds)
			if fastfail {
				return err
			}
			outerErr = errors.Join(outerErr, err)
		}
		for _, v := range addr {
			if v.IsValid() && (ipv6 && ipv4 ||
				((ipv4 && netool.IsIPv4(v)) || (ipv6 && netool.IsIPv6(v)))) {
				merged[v] = struct{}{}
			}
		}
		return nil
	}

	for _, ds := range datasources {
		if dualStack, isDualStack := ds.(DatasourceDualStack); isDualStack {
			if ipv4 {
				if err := addIPs(ds, "4", dualStack.IPv4); err != nil {
					return nil, err
				}
			}
			if ipv6 {
				if err := addIPs(ds, "6", dualStack.IPv6); err != nil {
					return nil, err
				}
			}
		} else {
			if err := addIPs(ds, "4/6", ds.IP); err != nil {
				return nil, err
			}
		}
	}

	return slices.Collect(maps.Keys(merged)), outerErr
}

type DatasourceError struct {
	Err       error
	IPVersion string
	Name      string
	Type      string
}

func newDatasourceError(err error, ipVersion string, ds Datasource) *DatasourceError {
	return &DatasourceError{
		Err:       err,
		IPVersion: ipVersion,
		Name:      ds.Name(),
		Type:      ds.Type(),
	}
}

func (e *DatasourceError) Error() string {
	return fmt.Sprintf("get ipv%s addresses from datasource(%s,%s) failed: %s",
		e.IPVersion, e.Type, e.Name, e.Err.Error())
}

func (e *DatasourceError) Unwrap() error {
	return e.Err
}

type DatasourceNotFoundError struct {
	*ManagedNotFoundError
}

func (e *DatasourceNotFoundError) Error() string {
	return fmt.Sprintf("datasource: %s", e.ManagedNotFoundError.Error())
}

func CreateDatasourceLogger(logger *zap.Logger, datasource Datasource) *zap.Logger {
	return logger.With(
		zap.String("datasource", datasource.Name()),
		zap.String("datasource_type", datasource.Type())).Named(
		"datasource")
}

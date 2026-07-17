package aliyun

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"

	"go.uber.org/zap"
)

const ProviderType = constpkg.ProviderTypeAliyun

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

var _ adapter.Provider = (*Aliyun)(nil)

type Aliyun struct {
	adapter.AbstractManagedType

	reconciler *ddnsx.Reconciler[Record]
}

func New(ctx context.Context, logger *zap.Logger, option options.AliyunProviderOption) (adapter.Provider, error) {
	if option.AccessKeyId == "" {
		return nil, fmt.Errorf("accessKeyId is empty")
	}
	if option.AccessKeySecret == "" {
		return nil, fmt.Errorf("accessKeySecret is empty")
	}

	_, _, httpClient, err := castoption.BuildHTTPClientFromScratch(
		logger, option.Connect, option.DNS, option.HTTP)
	if err != nil {
		return nil, err
	}

	client := NewClient(logger, httpClient,
		option.AccessKeyId, option.AccessKeySecret, "")
	observed := providerx.NewMetricsClientFromContext(
		ctx, option.Name, ProviderType, client,
	)
	return &Aliyun{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		reconciler:          ddnsx.NewReconciler(logger.Named("client"), observed),
	}, nil
}

func (a *Aliyun) Close() error { return nil }

func (a *Aliyun) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	return a.reconciler.Diff(ctx, domain, addr)
}

func (a *Aliyun) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	return a.reconciler.Update(ctx, domain, ttl, addr)
}

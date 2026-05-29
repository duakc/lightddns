package options_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/options"

	goyaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func init() {
	adapter.Register(adapter.DatasourceRegister, "http",
		func(ctx context.Context, opt options.HTTPDatasourceOption) (adapter.Datasource, error) {
			return nil, nil
		})
	adapter.Register(adapter.ProviderRegister, "cloudflare",
		func(ctx context.Context, opt options.CloudflareProviderOption) (adapter.Provider, error) {
			return nil, nil
		})
	adapter.Register(adapter.ServiceRegistry, "prometheus",
		func(ctx context.Context, opt options.PrometheusServiceOption) (adapter.Service, error) {
			return nil, nil
		})
}

func TestAutoNameInner(t *testing.T) {
	yamlBytes := []byte(`
providers:
  - type: cloudflare
    token: "test"
datasources:
  - type: http
    url: "https://ip.sb"
services:
  - type: prometheus
    enabled: true
    port: 9801
`)
	t.Run("badyaml.Unmarshal", func(t *testing.T) {
		var opt options.Options
		require.NoError(t, badyaml.Unmarshal(yamlBytes, &opt))
		checkInnerNames(t, &opt)
	})

	t.Run("goyaml.Decoder", func(t *testing.T) {
		var opt options.Options
		decoder := goyaml.NewDecoder(bytes.NewReader(yamlBytes), goyaml.DisallowUnknownField())
		require.NoError(t, decoder.Decode(&opt))
		checkInnerNames(t, &opt)
	})
}

func checkInnerNames(t *testing.T, opt *options.Options) {
	t.Helper()

	require.NotEmpty(t, opt.Datasources[0].Name, "ds outer Name empty")
	dsInner, ok := opt.Datasources[0].Option.(*options.HTTPDatasourceOption)
	require.True(t, ok)
	t.Logf("ds outer=%q inner=%q", opt.Datasources[0].Name, dsInner.Name)
	require.Equal(t, opt.Datasources[0].Name, dsInner.Name, "ds inner Name should match outer")

	require.NotEmpty(t, opt.Providers[0].Name, "provider outer Name empty")
	pInner, ok := opt.Providers[0].Option.(*options.CloudflareProviderOption)
	require.True(t, ok)
	t.Logf("provider outer=%q inner=%q", opt.Providers[0].Name, pInner.Name)
	require.Equal(t, opt.Providers[0].Name, pInner.Name, "provider inner Name should match outer")

	require.NotEmpty(t, opt.Services[0].Name, "service outer Name empty")
	sInner, ok := opt.Services[0].Option.(*options.PrometheusServiceOption)
	require.True(t, ok)
	t.Logf("service outer=%q inner=%q", opt.Services[0].Name, sInner.Name)
	require.Equal(t, opt.Services[0].Name, sInner.Name, "service inner Name should match outer")
}

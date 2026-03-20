package cloudflare

import (
	"context"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	CST "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
)

type cloudflare struct {
	name string

	client *internal.Client
}

func New(ctx context.Context, option options.OptionProviderCloudflare) (adapter.Provider, error) {

}

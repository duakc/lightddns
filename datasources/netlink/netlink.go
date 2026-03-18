package netlink

import (
	"context"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/options"
)

func New(ctx context.Context, option options.OptionDataSourceNetlink) (adapter.DataSource, error) {
	return newNetLink(ctx, option)
}

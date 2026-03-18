//go:build !linux

package netlink

import (
	"context"
	"fmt"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/options"
)

func newNetLink(ctx context.Context, option options.OptionDataSourceNetlink) (adapter.DataSource, error) {
	return nil, fmt.Errorf("netlink only support platform linux")
}

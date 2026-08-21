package build

import (
	"context"
	"flag"
	"runtime"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
)

func Run(ctx context.Context) {
	var (
		goos   string
		goarch string
	)

	flag.StringVar(&goos, "goos", "", "GOOS to build (e.g. linux); empty matches every OS")
	flag.StringVar(&goarch, "goarch", "", "GOARCH to build (e.g. amd64); empty matches every arch")
	flag.Parse()

	params := gobuild.DefaultParams()

	targets := target.FilterTargets(target.All(), goos, goarch)

	if len(targets) == 0 {
		common.Warnf("no target matches --os %q --arch %q (host GOOS=%s GOARCH=%s)",
			goos, goarch, runtime.GOOS, runtime.GOARCH)
		return
	}

	params.Qualified = params.Qualified || len(targets) > 1

	for _, tgt := range targets {
		if _, err := gobuild.Binary(ctx, tgt, params); err != nil {
			common.Fatalf("%s", err.Error())
		}
	}
}

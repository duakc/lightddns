package build

import (
	"context"
	"flag"
	"runtime"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
)

func Run(ctx context.Context) {
	params := gobuild.DefaultParams()
	var (
		tags    string
		verbose bool
		all     bool
		goos    string
		goarch  string
	)
	flag.StringVar(&params.Version, "version", "", "version (default: git tag or short hash)")
	flag.StringVar(&params.Branch, "branch", "", "branch (default: current git branch)")
	flag.StringVar(&params.WorkingDir, "workdir", params.WorkingDir, "package to build")
	flag.StringVar(&params.BinaryName, "binary", params.BinaryName, "binary name")
	flag.StringVar(&tags, "tags", "", "extra build tags (comma separated)")
	flag.StringVar(&goos, "os", "", "GOOS to build (e.g. linux); empty matches every OS")
	flag.StringVar(&goarch, "arch", "", "GOARCH to build (e.g. amd64); empty matches every arch")
	flag.BoolVar(&all, "all", false, "build every target (all OS/arch)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	common.Verbose = verbose
	if params.Version == "" {
		params.Version = gitver.Version(ctx)
	}
	if params.Branch == "" {
		params.Branch = gitver.Branch(ctx)
	}

	if tags != "" {
		params.ExtraTags = append(params.ExtraTags, strings.Split(tags, ",")...)
	}

	// Select targets: --os/--arch (or --all) filters the matrix; with no
	// selection, default to the host's baseline target only.
	hostOnly := !all && goos == "" && goarch == ""
	var targets []target.Target
	for _, t := range target.All() {
		matches := (goos == "" || t.GOOS == goos) && (goarch == "" || t.GOARCH == goarch)
		if hostOnly {
			matches = t.GOOS == runtime.GOOS && t.GOARCH == runtime.GOARCH
		}
		if !matches {
			continue
		}
		targets = append(targets, t)
		if hostOnly {
			break
		}
	}
	if len(targets) == 0 {
		common.Fatalf("no target matches --os %q --arch %q (host GOOS=%s GOARCH=%s)",
			goos, goarch, runtime.GOOS, runtime.GOARCH)
	}
	params.Qualified = len(targets) > 1

	for _, tgt := range targets {
		if _, err := gobuild.Binary(ctx, tgt, params); err != nil {
			common.Fatalf("%s", err.Error())
		}
	}
}

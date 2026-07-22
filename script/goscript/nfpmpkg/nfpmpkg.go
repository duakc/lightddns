// Package nfpmpkg is the nfpm-based builder for the systemd distro packages
// (deb, rpm, archlinux). It is the migration target for the hand-rolled deb/
// rpm/archlinux commands, which still exist unchanged; both paths coexist until
// nfpm is proven, then the old ones can be removed.
//
//	go run run.go nfpm --format deb|rpm|archlinux|all [--all]
//
// Each format's metadata, scripts and arch naming live in its own file
// (deb.go, rpm.go, archlinux.go); this file only parses flags, stages the
// shared assets, and dispatches. No format references another.
package nfpmpkg

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/nfpmbuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/ipk"
	_ "github.com/goreleaser/nfpm/v2/rpm"
)

// params carries the shared build inputs into each format builder.
type params struct {
	ctx          context.Context
	buildVersion string // stamps the binary (raw tag/hash)
	buildBranch  string
	buildAll     bool
	configPath   string // rendered /etc/lightddns.yaml
	manPath      string // gzipped man page
}

func Run(ctx context.Context) {
	var (
		format       string
		buildVersion string
		buildBranch  string
		buildAll     bool
		verbose      bool
	)
	flag.StringVar(&format, "format", "all", "package format: deb|rpm|archlinux|openwrt|all")
	flag.StringVar(&buildVersion, "version", "", "package version (default: git tag or short hash)")
	flag.StringVar(&buildBranch, "branch", "", "build branch (default: current git branch)")
	flag.BoolVar(&buildAll, "all", false, "build every shipped arch (default: host arch only)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()
	common.Verbose = verbose

	if buildVersion == "" {
		buildVersion = gitver.Version(ctx)
	}
	if buildBranch == "" {
		buildBranch = gitver.Branch(ctx)
	}

	// The config and man page are arch- and format-independent, so stage them
	// once and reuse across every package.
	stage := common.BuildDraftDir("nfpm", "shared")
	if err := os.RemoveAll(stage); err != nil {
		common.Fatalf("%s", err)
	}
	configPath, err := nfpmbuild.StageConfig(stage, nfpmbuild.SchemaURL(buildVersion))
	if err != nil {
		common.Fatalf("stage config: %s", err)
	}
	manPath, err := nfpmbuild.StageMan(stage)
	if err != nil {
		common.Fatalf("stage man: %s", err)
	}

	p := params{
		ctx:          ctx,
		buildVersion: buildVersion,
		buildBranch:  buildBranch,
		buildAll:     buildAll,
		configPath:   configPath,
		manPath:      manPath,
	}

	switch format {
	case "deb":
		buildDeb(p)
	case "rpm":
		buildRPM(p)
	case "archlinux", "arch":
		buildArch(p)
	case "openwrt":
		buildOpenWrt(p)
	case "all":
		buildDeb(p)
		buildRPM(p)
		buildArch(p)
		buildOpenWrt(p)
	default:
		common.Fatalf("unknown --format %q (want deb|rpm|archlinux|openwrt|all)", format)
	}
}

// compile builds tgt's binary into a per-format, per-arch staging dir and
// returns its path.
func (p params) compile(sub string, tgt target.Target) (string, error) {
	binDir := common.BuildDraftDir("nfpm", sub, tgt.BinaryName("lightddns"))
	if err := os.RemoveAll(binDir); err != nil {
		return "", err
	}
	return gobuild.Plain(p.ctx, tgt, binDir, p.buildVersion, p.buildBranch)
}

// sanitizeVersion strips a leading "v" and, using sep for a dash, keeps a
// digit-first version. deb/rpm order "~" as a pre-release; pacman uses "_".
func sanitizeVersion(v, sep string) string {
	v = strings.TrimPrefix(v, "v")
	if sep != "" {
		v = strings.ReplaceAll(v, "-", sep)
	}
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0" + orTilde(sep) + v
	}
	return v
}

func orTilde(sep string) string {
	if sep == "" {
		return "~"
	}
	return sep
}

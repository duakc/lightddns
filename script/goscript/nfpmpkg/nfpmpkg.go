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
	"runtime"
	"strings"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/goreleaser/nfpm/v2"

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/ipk"
	_ "github.com/goreleaser/nfpm/v2/rpm"
)

func Run(ctx context.Context) {
	var (
		goos   string
		goarch string

		format   string
		buildAll bool
	)
	flag.StringVar(&goos, "goos", runtime.GOOS, "target OS")
	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "target architecture")
	flag.StringVar(&format, "format", "all", "package format: deb|rpm|archlinux|openwrt|all")
	flag.BoolVar(&buildAll, "all", false, "build every shipped arch (default: host arch only)")
	flag.Parse()

	if format == "all" {
		buildDeb(p)
		buildRPM(p)
		buildArch(p)
		buildOpenWrt(p)
	}

	switch format {
	case packing.PackageDEB.String():
		buildDeb(p)
	case packing.PackageRPM.String():
		buildRPM(p)
	case packing.PackageArchLinux.String():
		buildArch(p)
	case packing.PackageAPK.String(), packing.PackageIPK.String():
		buildOpenWrt(p)
	default:
		common.Fatalf("unknown package format: %s", format)
	}
}

func sanitizeVersion(v, sep string) string {
	v = strings.TrimPrefix(v, "v")
	if sep != "" {
		v = strings.ReplaceAll(v, "-", sep)
	} else {
		sep = "~"
	}

	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0" + sep + v
	}
	return v
}

func BaseInfo(goarch, pkgVersion string) *nfpm.Info {
	return &nfpm.Info{
		Name:          constpkg.Project,
		Description:   constpkg.ProjectDescription,
		Homepage:      constpkg.DocsURL,
		License:       constpkg.LICENSE,
		VersionSchema: "semver",
		Maintainer:    "young <young@qeee.net>",

		Arch:    goarch,
		Version: pkgVersion,
	}
}

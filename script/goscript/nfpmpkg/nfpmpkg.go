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
	"errors"
	"flag"
	"os"
	"runtime"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/duakc/mt/services/filehelper"

	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/ipk"
	_ "github.com/goreleaser/nfpm/v2/rpm"
)

const nfpmGOOS = "linux"

func Run(ctx context.Context) {
	var (
		goarch string

		format   string
		buildAll bool
	)
	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "target architecture")
	flag.StringVar(&format, "format", "", "package format: deb|rpm|archlinux|ipk|openwrt.apk|alpine.apk|openwrt")
	flag.BoolVar(&buildAll, "all", false, "build every package format")
	flag.Parse()

	if packing.PackageDEB.String() == format || buildAll {
		debTargets := target.DEBTargets(target.All(), nfpmGOOS, goarch)
		buildDeb(ctx, debTargets)
	}

	if packing.PackageRPM.String() == format || buildAll {
		rpmTargets := target.RPMTargets(target.All(), nfpmGOOS, goarch)
		buildRPM(ctx, rpmTargets)
	}

	if packing.PackageArchLinux.String() == format || buildAll {
		archLinuxTargets := target.ArchLinuxTargets(target.All(), nfpmGOOS, goarch)
		buildArchLinux(ctx, archLinuxTargets)
	}

	if packing.PackageAlpineAPK.String() == format || buildAll {
		apkTargets := target.AlpineAPKTargets(target.All(), nfpmGOOS, goarch)
		buildAlpineAPK(ctx, apkTargets)
	}

	if format == "openwrt" || buildAll {
		openWrtTargets := target.OpenWrtTargets(target.All(), nfpmGOOS, goarch)
		buildOpenWrtIPK(ctx, openWrtTargets)
		buildOpenWrtAPK(ctx, openWrtTargets)
	}

	if packing.PackageIPK.String() == format {
		openWrtTargets := target.OpenWrtTargets(target.All(), nfpmGOOS, goarch)
		buildOpenWrtIPK(ctx, openWrtTargets)
	}

	if packing.PackageOpenWrtAPK.String() == format {
		openWrtTargets := target.OpenWrtTargets(target.All(), nfpmGOOS, goarch)
		buildOpenWrtAPK(ctx, openWrtTargets)
	}
}

func BaseInfo(goarch string) *nfpm.Info {
	return &nfpm.Info{
		Name:          constpkg.Project,
		Description:   constpkg.ProjectDescription,
		Homepage:      constpkg.DocsURL,
		License:       constpkg.LICENSE,
		VersionSchema: "semver",
		Maintainer:    "young <young@qeee.net>",
		Priority:      "optional",

		Arch:    goarch,
		Version: gitver.Version(context.Background()),
	}
}

func writeToFile(
	fh filehelper.Helper,
	packageName string,
	packageType packing.PackageType,
	info *nfpm.Info,
) (*os.File, error) {
	packager, err := nfpm.Get(packageType.Nfpm())
	if err != nil {
		return nil, err
	}
	info = nfpm.WithDefaults(info)
	if err := nfpm.Validate(info); err != nil {
		return nil, err
	}
	file, err := fh.Create(packageName)
	if err != nil {
		return nil, err
	}

	err = packager.Package(info, file)
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return file, nil
}

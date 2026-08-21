package nfpmpkg

import (
	"context"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/buildinfo"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"

	"github.com/Masterminds/semver/v3"
	"github.com/goreleaser/nfpm/v2"
)

const openWrtReleaseVersion = "1"

func buildOpenWrtIPK(ctx context.Context, targets []target.Target) {
	if len(targets) == 0 {
		common.Warnf("skip openwrt ipk package build: no matching targets")
		return
	}

	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "openwrt", "ipk"))
	defer outputDir.Close()

	baseContents := (&FileContents{}).
		AddConfig(SchemaURL(gitver.LocatableVersion(ctx))).
		AddEnvFile().
		AddOpenWrtInit()

	built := 0
	for _, tgt := range targets {
		if !tgt.OpenWrt {
			continue
		}

		binary, err := BuildBinary(ctx, tgt, packing.PackageIPK)
		if err != nil {
			common.Fatalf("%s", err)
		}

		for _, arch := range tgt.OpenWrtArch {
			contents := baseContents.Copy().AddBinaryPath(binary)

			info := BaseInfo(tgt.GOARCH)
			info.Overridables = nfpm.Overridables{
				Contents: contents.Contents,
				Scripts: nfpm.Scripts{
					PostInstall: common.ReleaseDir("openwrt", "postinst"),
					PreRemove:   common.ReleaseDir("openwrt", "prerm"),
				},
				IPK: nfpm.IPK{
					Arch: arch,
					Fields: map[string]string{
						"Section": "net",
					},
				},
			}
			openWrtVersion(info, buildinfo.Semver())

			file, err := writeToFile(outputDir, openWrtName(info, packing.PackageIPK, arch), packing.PackageIPK, info)
			if err != nil {
				common.Fatalf("pack openwrt ipk %s", err)
			}

			built++
			common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
			_ = file.Close()
		}
	}
	common.Infof("done, built %d openwrt ipk(s)", built)
}

func buildOpenWrtAPK(ctx context.Context, targets []target.Target) {
	if len(targets) == 0 {
		common.Warnf("skip openwrt apk package build: no matching targets")
		return
	}

	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "openwrt", "apk"))
	defer outputDir.Close()

	baseContents := (&FileContents{}).
		AddConfig(SchemaURL(gitver.LocatableVersion(ctx))).
		AddEnvFile().
		AddOpenWrtInit()

	built := 0
	for _, tgt := range targets {
		if !tgt.OpenWrt {
			continue
		}

		binary, err := BuildBinary(ctx, tgt, packing.PackageOpenWrtAPK)
		if err != nil {
			common.Fatalf("%s", err)
		}

		for _, arch := range tgt.OpenWrtArch {
			contents := baseContents.Copy().AddBinaryPath(binary)

			info := BaseInfo(tgt.GOARCH)
			info.Overridables = nfpm.Overridables{
				Contents: contents.Contents,
				Scripts: nfpm.Scripts{
					PostInstall: common.ReleaseDir("openwrt", "postinst"),
					PreRemove:   common.ReleaseDir("openwrt", "prerm"),
				},
				APK: nfpm.APK{
					Arch: arch,
				},
			}
			openWrtVersion(info, buildinfo.Semver())

			file, err := writeToFile(outputDir, openWrtName(info, packing.PackageOpenWrtAPK, arch), packing.PackageOpenWrtAPK, info)
			if err != nil {
				common.Fatalf("pack openwrt apk %s", err)
			}

			built++
			common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
			_ = file.Close()
		}
	}
	common.Infof("done, built %d openwrt apk(s)", built)
}

func openWrtName(info *nfpm.Info, packageType packing.PackageType, arch string) string {
	return target.QualifyName(
		constpkg.Project, info.Version, info.Release, arch) +
		"." + packageType.Ext()
}

func openWrtVersion(info *nfpm.Info, version *semver.Version) {
	info.Version = version.String()
	info.Release = openWrtReleaseVersion
}

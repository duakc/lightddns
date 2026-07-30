package nfpmpkg

import (
	"context"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"

	"github.com/Masterminds/semver/v3"
	"github.com/goreleaser/nfpm/v2"
)

const alpineReleaseVersion = "1"

func buildAlpineAPK(ctx context.Context, targets []target.Target, baseContents *FileContents) {
	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "alpine", "apk"))
	defer outputDir.Close()

	built := 0
	for _, tgt := range targets {
		if !tgt.Alpine {
			continue
		}

		contents := baseContents.Copy().
			AddBinary(ctx, tgt, packing.PackageAlpineAPK)

		info := BaseInfo(tgt.GOARCH)
		info.Overridables = nfpm.Overridables{
			Contents: contents.Contents,
			Scripts: nfpm.Scripts{
				PostInstall: common.ReleaseDir("alpine", "postinstall"),
				PreRemove:   common.ReleaseDir("alpine", "preremove"),
			},
			APK: nfpm.APK{
				Arch: tgt.AlpineArch,
			},
		}
		alpineVersion(info, gitver.Semver(ctx))

		file, err := writeToFile(outputDir, alpineAPKName(info, tgt), packing.PackageAlpineAPK, info)
		if err != nil {
			common.Fatalf("pack alpine apk %s", err)
		}

		built++
		common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
		_ = file.Close()
	}
	common.Infof("done, built %d alpine apk(s)", built)
}

func alpineAPKName(info *nfpm.Info, tgt target.Target) string {
	return target.QualifyName(
		constpkg.Project, info.Version, info.Release, tgt.AlpineArch) +
		"." + packing.PackageAlpineAPK.Ext()
}

func alpineVersion(info *nfpm.Info, version *semver.Version) {
	info.Version = version.String()
	info.Release = alpineReleaseVersion
}

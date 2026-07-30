package nfpmpkg

import (
	"context"

	"github.com/Masterminds/semver/v3"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"
	"github.com/goreleaser/nfpm/v2"
)

const (
	rpmReleaseVersion = "1"
)

func buildRPM(ctx context.Context, targets []target.Target, baseContents *FileContents) {
	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "rpm"))
	defer outputDir.Close()

	for _, tgt := range targets {
		if !tgt.RPM {
			continue
		}
		contents := baseContents.Copy().
			AddBinary(ctx, tgt, packing.PackageRPM)

		info := BaseInfo(tgt.GOARCH)

		info.Overridables = nfpm.Overridables{
			Depends:  []string{"shadow-utils"},
			Contents: contents.Contents,
			Scripts: nfpm.Scripts{
				PreInstall:  common.ReleaseDir("rpm", "scripts", "pre"),
				PostInstall: common.ReleaseDir("rpm", "scripts", "post"),
				PreRemove:   common.ReleaseDir("rpm", "scripts", "preun"),
				PostRemove:  common.ReleaseDir("rpm", "scripts", "postun"),
			},
			RPM: nfpm.RPM{
				Arch: tgt.RPMArchName(),
			},
		}

		rpmVersion(info, gitver.Semver(ctx))

		file, err := writeToFile(outputDir, rpmName(info, tgt), packing.PackageRPM, info)
		if err != nil {
			common.Fatalf("pack rpm %s", err)
		}

		common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
		_ = file.Close()
	}
	common.Infof("done, built %d rpm(s)", len(targets))
}

func rpmName(info *nfpm.Info, tgt target.Target) string {
	return target.QualifyName(
		constpkg.Project, info.Version, info.Release, info.RPM.Arch) +
		"." + packing.PackageRPM.Ext()
}

func rpmVersion(info *nfpm.Info, version *semver.Version) {
	info.Version = version.String()
	info.Release = rpmReleaseVersion
	// RPM doesn't accept Prerelease .
	// info.Prerelease = version.Prerelease()
}

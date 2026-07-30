package nfpmpkg

import (
	"cmp"
	"context"
	"fmt"

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

const (
	debReleaseVersion = "1"
)

func buildDeb(ctx context.Context, targets []target.Target, baseContents *FileContents) {
	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "deb"))
	defer outputDir.Close()

	for _, tgt := range targets {
		if !tgt.DEB {
			continue
		}
		contents := baseContents.Copy().
			AddBinary(ctx, tgt, packing.PackageDEB)

		info := BaseInfo(tgt.GOARCH)

		info.Overridables = nfpm.Overridables{
			Depends:  []string{"adduser"},
			Contents: contents.Contents,
			Scripts: nfpm.Scripts{
				PostInstall: common.ReleaseDir("deb", "scripts", "postinst"),
				PreRemove:   common.ReleaseDir("deb", "scripts", "prerm"),
				PostRemove:  common.ReleaseDir("deb", "scripts", "postrm"),
			},
			Deb: nfpm.Deb{
				Arch:        tgt.DEBArch,
				ArchVariant: tgt.DEBArchVariant,
			},
		}

		debVersion(info, gitver.Semver(ctx))

		file, err := writeToFile(outputDir, debName(info, tgt), packing.PackageDEB, info)
		if err != nil {
			common.Fatalf("pack deb %s", err)
		}

		common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
		_ = file.Close()
	}
	common.Infof("done, built %d deb(s)", len(targets))
}

func debName(info *nfpm.Info, tgt target.Target) string {
	baseNames := []string{
		constpkg.Project, info.Version, info.Prerelease, info.Release,
		cmp.Or(info.Deb.ArchVariant, info.Deb.Arch, tgt.GOARCH),
	}

	return target.QualifyName(baseNames...) +
		"." + packing.PackageDEB.Ext()
}

func debVersion(info *nfpm.Info, version *semver.Version) {
	info.Version = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(),
		version.Patch())
	info.Release = debReleaseVersion
	info.Prerelease = version.Prerelease()
}

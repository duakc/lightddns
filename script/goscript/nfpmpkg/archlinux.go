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

const archLinuxReleaseVersion = "1"

func buildArchLinux(ctx context.Context, targets []target.Target) {
	outputDir := filehelper.MustNew(common.BuildDir("nfpm", "archlinux"))
	defer outputDir.Close()

	baseContents := (&FileContents{}).
		AddConfig(SchemaURL(gitver.LocatableVersion(ctx))).
		AddEnvFile().
		AddSystemdService().
		AddSystemdSysUsers().
		AddSystemdTmpFiles().
		AddMan()

	built := 0
	for _, tgt := range targets {
		if !tgt.ArchLinux {
			continue
		}

		contents := baseContents.Copy().
			AddBinary(ctx, tgt, packing.PackageArchLinux)

		info := BaseInfo(tgt.GOARCH)
		info.Overridables = nfpm.Overridables{
			Contents: contents.Contents,
			Scripts: nfpm.Scripts{
				PostInstall: common.ReleaseDir("archlinux", "scripts", "post_install"),
				PreRemove:   common.ReleaseDir("archlinux", "scripts", "pre_remove"),
				PostRemove:  common.ReleaseDir("archlinux", "scripts", "post_remove"),
			},
			ArchLinux: nfpm.ArchLinux{
				Arch:     tgt.ArchLinuxArch,
				Packager: "young <young@qeee.net>",
				Scripts: nfpm.ArchLinuxScripts{
					PostUpgrade: common.ReleaseDir("archlinux", "scripts", "post_upgrade"),
				},
			},
		}
		archLinuxVersion(info, gitver.Semver(ctx))

		file, err := writeToFile(outputDir, archLinuxName(info, tgt), packing.PackageArchLinux, info)
		if err != nil {
			common.Fatalf("pack archlinux %s", err)
		}

		built++
		common.Infof("built %s (size %d)", file.Name(), mt.Must(file.Stat()).Size())
		_ = file.Close()
	}
	common.Infof("done, built %d archlinux package(s)", built)
}

func archLinuxName(info *nfpm.Info, tgt target.Target) string {
	return target.QualifyName(
		constpkg.Project, info.Version, info.Release, tgt.ArchLinuxArch) +
		"." + packing.PackageArchLinux.Ext()
}

func archLinuxVersion(info *nfpm.Info, version *semver.Version) {
	info.Version = version.String()
	info.Release = archLinuxReleaseVersion
}

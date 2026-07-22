package nfpmpkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/nfpmbuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2"
)

// rpmRelease is the RPM Release field; pinned so the filename is deterministic.
const rpmRelease = "1"

func buildRPM(p params) {
	outputDir := common.BuildDir("nfpm", "rpm")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		common.Fatalf("%s", err)
	}
	targets := target.RPMTargets(p.buildAll)
	if len(targets) == 0 {
		common.Fatalf("no rpm built for host arch (try --all)")
	}
	// rpm forbids "-" in Version; "~" orders as a pre-release.
	pkgVersion := sanitizeVersion(p.buildVersion, "~")

	for _, tgt := range targets {
		binPath, err := p.compile("rpm", tgt)
		if err != nil {
			common.Fatalf("compile: %s", err)
		}
		rpmArch := tgt.RPMArchName()

		info := &nfpm.Info{
			Name:          "lightddns",
			Arch:          tgt.GOARCH,
			Version:       pkgVersion,
			VersionSchema: "none",
			Release:       rpmRelease,
			Maintainer:    "duakc <young@qeee.net>",
			Description:   "Lightweight dynamic DNS (DDNS) updater",
			Homepage:      "https://lightddns.duaky.com",
			License:       "GPL-2.0-only",
			Overridables: nfpm.Overridables{
				Depends:  []string{"shadow-utils"},
				Contents: append(nfpmbuild.SystemdContents(p.configPath, p.manPath), nfpmbuild.Binary(binPath)),
				Scripts: nfpm.Scripts{
					PreInstall:  common.ReleaseDir("rpm", "scripts", "pre"),
					PostInstall: common.ReleaseDir("rpm", "scripts", "post"),
					PreRemove:   common.ReleaseDir("rpm", "scripts", "preun"),
					PostRemove:  common.ReleaseDir("rpm", "scripts", "postun"),
				},
			},
		}
		info.RPM.Arch = rpmArch

		out := filepath.Join(outputDir, fmt.Sprintf("lightddns-%s-%s.%s.rpm", pkgVersion, rpmRelease, rpmArch))
		if err := nfpmbuild.WriteTo(info, "rpm", out); err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", out)
	}
	common.Infof("done, built %d rpm(s)", len(targets))
}

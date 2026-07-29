package nfpmpkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/nfpmbuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2"
)

func buildDeb(ctx context.Context, targets target.Target, baseContents *FileContents) {
	outputDir := common.BuildDir("nfpm", "deb")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		common.Fatalf("%s", err)
	}
	targets := target.DEBTargets(p.buildAll)
	if len(targets) == 0 {
		common.Fatalf("no deb built for host arch (try --all)")
	}
	pkgVersion := sanitizeVersion(p.buildVersion, "")

	for _, tgt := range targets {
		binPath, err := p.compileBinary("deb", tgt)
		if err != nil {
			common.Fatalf("compile: %s", err)
		}
		debArch := tgt.DEBArchName()

		info := &nfpm.Info{
			Name:          "lightddns",
			Arch:          tgt.GOARCH,
			Version:       pkgVersion,
			VersionSchema: "none",
			Section:       "net",
			Priority:      "optional",
			Maintainer:    "duakc <young@qeee.net>",
			Description:   "Lightweight dynamic DNS (DDNS) updater",
			Homepage:      "https://lightddns.duaky.com",
			License:       "GPL-2.0",
			Overridables: nfpm.Overridables{
				Depends:  []string{"adduser"},
				Contents: append(nfpmbuild.ContentsSystemdService(p.configPath, p.manPath), nfpmbuild.Binary(binPath)),
				Scripts: nfpm.Scripts{
					PostInstall: common.ReleaseDir("deb", "pkgroot", "DEBIAN", "postinst"),
					PreRemove:   common.ReleaseDir("deb", "pkgroot", "DEBIAN", "prerm"),
					PostRemove:  common.ReleaseDir("deb", "pkgroot", "DEBIAN", "postrm"),
				},
			},
		}
		info.Deb.Arch = debArch

		out := filepath.Join(outputDir, fmt.Sprintf("lightddns_%s_%s.deb", pkgVersion, debArch))
		if err := nfpmbuild.WriteTo(info, "deb", out); err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", out)
	}
	common.Infof("done, built %d deb(s)", len(targets))
}

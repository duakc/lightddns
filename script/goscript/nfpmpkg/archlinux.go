package nfpmpkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/nfpmbuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
)

const archRelease = "1"

func buildArch(p params) {
	outputDir := common.BuildDir("nfpm", "archlinux")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		common.Fatalf("%s", err)
	}
	targets := target.ArchLinuxTargets(p.buildAll)
	if len(targets) == 0 {
		common.Fatalf("no Arch Linux package built for host arch (try --all)")
	}
	// pacman forbids "-" in pkgver; it uses "_".
	pkgVersion := sanitizeVersion(p.buildVersion, "_")

	for _, tgt := range targets {
		arch := tgt.ArchLinuxArchName()
		if arch == "" {
			common.Fatalf("unsupported Arch Linux target: GOOS=%s GOARCH=%s", tgt.GOOS, tgt.GOARCH)
		}
		binPath, err := p.compile("archlinux", tgt)
		if err != nil {
			common.Fatalf("compile: %s", err)
		}

		contents := append(nfpmbuild.SystemdContents(p.configPath, p.manPath), nfpmbuild.Binary(binPath))
		contents = append(contents,
			&files.Content{Source: common.ReleaseDir("systemd", "sysusers", "lightddns.conf"), Destination: "/usr/lib/sysusers.d/lightddns.conf", FileInfo: &files.ContentFileInfo{Mode: 0o644}},
			&files.Content{Source: common.ReleaseDir("systemd", "tmpfiles", "lightddns.conf"), Destination: "/usr/lib/tmpfiles.d/lightddns.conf", FileInfo: &files.ContentFileInfo{Mode: 0o644}},
		)

		info := &nfpm.Info{
			Name:          "lightddns",
			Arch:          tgt.GOARCH,
			Version:       pkgVersion,
			VersionSchema: "none",
			Release:       archRelease,
			Maintainer:    "duakc <young@qeee.net>",
			Description:   "Lightweight dynamic DNS (DDNS) updater",
			Homepage:      "https://lightddns.duaky.com",
			License:       "GPL-2.0-only",
			Overridables: nfpm.Overridables{
				Depends:  []string{"systemd"},
				Contents: contents,
				Scripts: nfpm.Scripts{
					PostInstall: common.ReleaseDir("archlinux", "scripts", "post_install"),
					PreRemove:   common.ReleaseDir("archlinux", "scripts", "pre_remove"),
					PostRemove:  common.ReleaseDir("archlinux", "scripts", "post_remove"),
				},
			},
		}
		info.ArchLinux.Arch = arch
		info.ArchLinux.Scripts.PostUpgrade = common.ReleaseDir("archlinux", "scripts", "post_upgrade")

		out := filepath.Join(outputDir, fmt.Sprintf("lightddns-%s-%s-%s.pkg.tar.zst", pkgVersion, archRelease, arch))
		if err := nfpmbuild.WriteTo(info, "archlinux", out); err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", out)
	}
	common.Infof("done, built %d Arch Linux package(s)", len(targets))
}

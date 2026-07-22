package nfpmpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/nfpmbuild"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
)

// buildOpenWrt emits both .ipk (opkg) and *.openwrt.apk for every OpenWrt
// subtarget label. Unlike the systemd formats it ships a procd init script (no
// systemd unit, no man page) and reuses one binary across all labels of a Go
// build variant.
func buildOpenWrt(p params) {
	outputDir := common.BuildDir("nfpm", "openwrt")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		common.Fatalf("%s", err)
	}
	targets := target.OpenWrtTargets(p.buildAll)
	if len(targets) == 0 {
		common.Fatalf("no OpenWrt package for GOARCH=%s (try --all)", runtime.GOARCH)
	}
	pkgVersion := sanitizeVersion(p.buildVersion, "")

	built := 0
	for _, owt := range targets {
		// One binary per variant; keyed by the first label since BinaryName
		// collides across GO386/GOMIPS flavours.
		binDir := common.BuildDraftDir("nfpm", "openwrt", owt.Archs[0])
		if err := os.RemoveAll(binDir); err != nil {
			common.Fatalf("%s", err)
		}
		binPath, err := gobuild.Plain(p.ctx, owt.Target, binDir, p.buildVersion, p.buildBranch)
		if err != nil {
			common.Fatalf("compile %s: %s", owt.Archs[0], err)
		}
		common.Infof("GOARCH=%s -> %s", owt.GOARCH, strings.Join(owt.Archs, " "))

		for _, arch := range owt.Archs {
			for _, format := range []string{"ipk", "apk"} {
				info := &nfpm.Info{
					Name:          "lightddns",
					Arch:          owt.GOARCH,
					Version:       pkgVersion,
					VersionSchema: "none",
					Section:       "net",
					Maintainer:    "duakc <young@qeee.net>",
					Description:   "Lightweight dynamic DNS (DDNS) updater",
					Homepage:      "https://lightddns.duaky.com",
					License:       "GPL-2.0",
					Overridables: nfpm.Overridables{
						Contents: files.Contents{
							nfpmbuild.Binary(binPath),
							{Source: p.configPath, Destination: "/etc/lightddns.yaml", Type: "config|noreplace", FileInfo: &files.ContentFileInfo{Mode: 0o600}},
							{Source: common.ReleaseDir("openwrt", "lightddns.init"), Destination: "/etc/init.d/lightddns", FileInfo: &files.ContentFileInfo{Mode: 0o755}},
						},
						Scripts: nfpm.Scripts{
							PostInstall: common.ReleaseDir("openwrt", "postinst"),
							PreRemove:   common.ReleaseDir("openwrt", "prerm"),
						},
					},
				}
				// OpenWrt uses the same subtarget label for both opkg and apk.
				switch format {
				case "ipk":
					info.IPK.Arch = arch
				case "apk":
					info.APK.Arch = arch
				}

				out := filepath.Join(outputDir, fmt.Sprintf("lightddns_%s_%s.%s", pkgVersion, arch, openwrtExt(format)))
				if err := nfpmbuild.WriteTo(info, format, out); err != nil {
					common.Fatalf("package %s %s: %s", format, arch, err)
				}
				common.Infof("built %s", out)
				built++
			}
		}
	}
	common.Infof("done, built %d OpenWrt package(s)", built)
}

// openwrtExt distinguishes the OpenWrt apk (*.openwrt.apk) from the Alpine one.
func openwrtExt(format string) string {
	if format == "apk" {
		return "openwrt.apk"
	}
	return "ipk"
}

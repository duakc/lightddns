// Package openwrt builds OpenWrt packages: both .ipk (opkg, OpenWrt < 24.10)
// and .apk (OpenWrt >= 24.10, written as *.openwrt.apk to distinguish it from
// the Alpine flavour). Unlike deb/rpm it needs no native tool - nfpm assembles
// the archives in pure Go - and it ships a procd init script instead of systemd
// units. One binary is compiled per Go build variant and republished under every
// OpenWrt subtarget arch label that shares it.
package openwrt

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"
	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/ipk"
)

const binaryBase = constpkg.Project

const subSchemaURL = "SCHEMA_URL"

var releaseDirFileHelper = mt.Must(filehelper.New(common.ReleaseDir()))

var (
	outputDir = common.BuildDir("openwrt")

	buildAll     bool
	buildVersion string
	buildBranch  string
	verbose      bool

	pkgVersion string
	schemaURL  string
)

func Run(ctx context.Context) {
	flag.StringVar(&buildVersion, "version", "", "package version (default: git tag or short hash)")
	flag.StringVar(&buildBranch, "branch", "", "build branch (default: current git branch)")
	flag.BoolVar(&buildAll, "all", false, "build every OpenWrt arch (default: host GOARCH only)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()
	common.Verbose = verbose

	if buildVersion == "" {
		buildVersion = gitver.Version(ctx)
	}
	if buildBranch == "" {
		buildBranch = gitver.Branch(ctx)
	}
	pkgVersion = sanitizeVersion(buildVersion)
	schemaURL = fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/release/schema.json",
		constpkg.Repo, buildVersion)

	targets := target.OpenWrtTargets(buildAll)
	if len(targets) == 0 {
		common.Fatalf("no OpenWrt package for GOARCH=%s (try --all)", runtime.GOARCH)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		common.Fatalf("%s", err)
	}

	// The config is arch-independent (only the schema URL is substituted), so
	// render it once and reuse it for every package.
	configPath, err := renderConfig()
	if err != nil {
		common.Fatalf("render config: %s", err)
	}

	built := 0
	for _, owt := range targets {
		binPath, err := compile(ctx, owt)
		if err != nil {
			common.Fatalf("compile %s: %s", owt.Archs[0], err)
		}
		common.Infof("GOARCH=%s -> %s", owt.GOARCH, strings.Join(owt.Archs, " "))

		for _, arch := range owt.Archs {
			for _, format := range []string{"ipk", "apk"} {
				out, err := pack(format, owt.GOARCH, arch, binPath, configPath)
				if err != nil {
					common.Fatalf("package %s %s: %s", format, arch, err)
				}
				common.Infof("built %s", out)
				built++
			}
		}
	}
	common.Infof("done, built %d package(s)", built)
}

// compile builds owt's binary once into a per-variant staging dir and returns
// its path. The first arch label keys the dir since it is unique per variant
// (BinaryName alone collides across GO386/GOMIPS flavours).
func compile(ctx context.Context, owt target.OpenWrtTarget) (string, error) {
	binDir := common.BuildDraftDir("openwrt", owt.Archs[0], "bin")
	if err := os.RemoveAll(binDir); err != nil {
		return "", err
	}
	return gobuild.Plain(ctx, owt.Target, binDir, buildVersion, buildBranch)
}

// renderConfig substitutes the schema URL into the example config and returns
// the rendered file path.
func renderConfig() (string, error) {
	stage := common.BuildDraftDir("openwrt", "root")
	if err := os.RemoveAll(stage); err != nil {
		return "", err
	}
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return "", err
	}
	stageFileHelper, err := filehelper.New(stage)
	if err != nil {
		return "", err
	}
	defer stageFileHelper.Close()

	fileList := []packing.File{
		{FS: releaseDirFileHelper, From: "example/lightddns.yaml", To: "lightddns.yaml", Mode: 0o600, SubSetVec: packing.SubSetVec{subSchemaURL}},
	}
	if err := packing.ProcessAll(stageFileHelper, fileList, packing.SubSet{subSchemaURL: schemaURL}); err != nil {
		return "", err
	}
	return filepath.Join(stage, "lightddns.yaml"), nil
}

// pack emits one package. goArch is the Go arch (for nfpm's defaults); arch is
// the OpenWrt subtarget label stamped into the package and filename.
func pack(format, goArch, arch, binPath, configPath string) (string, error) {
	info := &nfpm.Info{
		Name:          binaryBase,
		Arch:          goArch,
		Version:       pkgVersion,
		VersionSchema: "none",
		Section:       "net",
		Maintainer:    "young <young@qeee.net>",
		Description:   "Lightweight dynamic DNS (DDNS) updater",
		Homepage:      "https://lightddns.duaky.com",
		License:       "GPL-2.0",
		Overridables: nfpm.Overridables{
			Contents: files.Contents{
				{Source: binPath, Destination: "/usr/bin/lightddns", FileInfo: &files.ContentFileInfo{Mode: 0o755}},
				{Source: configPath, Destination: "/etc/lightddns.yaml", Type: "config", FileInfo: &files.ContentFileInfo{Mode: 0o600}},
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

	pkgr, err := nfpm.Get(format)
	if err != nil {
		return "", err
	}
	info = nfpm.WithDefaults(info)
	if err := nfpm.Validate(info); err != nil {
		return "", err
	}

	out := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%s.%s", binaryBase, pkgVersion, arch, ext(format)))
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := pkgr.Package(info, f); err != nil {
		return "", err
	}
	return out, nil
}

// ext distinguishes the OpenWrt apk (*.openwrt.apk) from the future Alpine one.
func ext(format string) string {
	if format == "apk" {
		return "openwrt.apk"
	}
	return "ipk"
}

// sanitizeVersion strips a leading "v" and keeps a digit-first version.
func sanitizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0~" + v
	}
	return v
}

// Package archlinux builds Arch Linux packages. Go prepares a makepkg workdir
// from the shared release assets, compiles the target binary into srcdir, renders
// PKGBUILD placeholders, then shells out to makepkg.
package archlinux

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
)

const (
	binaryBase  = constpkg.Project
	packageName = binaryBase + "-bin"
)

const (
	subVersionSubSet   = "VERSION"
	subSchemaURLSubSet = "SCHEMA_URL"
	subArchSubSet      = "ARCH"
	packageNameSubSet  = "PACKAGE_NAME"
)

var (
	releaseDirFileHelper = mt.Must(filehelper.New(common.ReleaseDir()))
	archlinuxFileHelper  = mt.Must(filehelper.New(common.ReleaseDir("archlinux")))
)

var packingFileList = []packing.File{
	{FS: archlinuxFileHelper, From: "PKGBUILD", To: "PKGBUILD", Mode: 0o644, SubSetVec: packing.SubSetVec{packageNameSubSet, subVersionSubSet, subArchSubSet}},
	{FS: releaseDirFileHelper, From: "example/lightddns.yaml", To: "pkgroot/etc/lightddns.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURLSubSet}},
	{FS: releaseDirFileHelper, From: "example/lightddns.yaml", To: "pkgroot/etc/lightddns.d/example.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURLSubSet}},
	{FS: releaseDirFileHelper, From: "example/environment", To: "pkgroot/etc/default/lightddns", Mode: 0o640},
	{FS: releaseDirFileHelper, From: "systemd/lightddns.service", To: "pkgroot/usr/lib/systemd/system/lightddns.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "systemd/lightddns@.service", To: "pkgroot/usr/lib/systemd/system/lightddns@.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "man/lightddns.1", To: "pkgroot/usr/share/man/man1/lightddns.1.gz", Mode: 0o644, Gzip: true},
	{FS: releaseDirFileHelper, From: "systemd/sysusers/lightddns.conf", To: "pkgroot/usr/lib/sysusers.d/lightddns.conf", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "systemd/tmpfiles/lightddns.conf", To: "pkgroot/usr/lib/tmpfiles.d/lightddns.conf", Mode: 0o644},
	{FS: archlinuxFileHelper, From: "lightddns.install", To: "lightddns.install", Mode: 0o644},
}

var (
	outputDir = common.BuildDir("archlinux")

	buildAll     bool
	buildVersion string
	buildBranch  string

	verbose bool

	pkgVersion string
	schemaURL  string
)

func Run(ctx context.Context) {
	flag.StringVar(&buildVersion, "version", "", "package version (default: git tag or short hash)")
	flag.StringVar(&buildBranch, "branch", "", "build branch (default: current git branch)")
	flag.BoolVar(&buildAll, "all", false, "build every shipped arch (Arch Linux currently ships x86_64 only)")
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

	built := 0
	for _, tgt := range target.All() {
		if !tgt.ArchLinux || (!buildAll && (runtime.GOARCH != tgt.GOARCH || runtime.GOOS != tgt.GOOS)) {
			continue
		}

		common.Infof("GOOS=%s GOARCH=%s TARGET_GOOS=%s TARGET_GOARCH=%s",
			runtime.GOOS, runtime.GOARCH, tgt.GOOS, tgt.GOARCH)

		pkg, err := pack(ctx, tgt)
		if err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", pkg)
		built++
	}

	if built == 0 {
		common.Fatalf("no Arch Linux package built with GOARCH=%s GOOS=%s", runtime.GOARCH, runtime.GOOS)
	}
	common.Infof("done, built %d package(s)", built)
}

func pack(ctx context.Context, tgt target.Target) (string, error) {
	arch := tgt.ArchLinuxArchName()
	if arch == "" {
		return "", fmt.Errorf("unsupported Arch Linux target: GOOS=%s GOARCH=%s", tgt.GOOS, tgt.GOARCH)
	}

	workdir := common.BuildDraftDir("archlinux", arch)
	if err := os.RemoveAll(workdir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return "", err
	}

	workdirFileHelper, err := filehelper.New(workdir)
	if err != nil {
		return "", err
	}
	defer workdirFileHelper.Close()

	subSet := packing.SubSet{
		packageNameSubSet:  packageName,
		subVersionSubSet:   pkgVersion,
		subSchemaURLSubSet: schemaURL,
		subArchSubSet:      arch,
	}
	if err := packing.ProcessAll(workdirFileHelper, packingFileList, subSet); err != nil {
		return "", err
	}
	if _, err := gobuild.Plain(ctx, tgt, filepath.Join(workdir, "pkgroot", "usr", "bin"), buildVersion, buildBranch); err != nil {
		return "", err
	}

	if err := common.CommandStream(ctx, common.Cmd{
		Name: "makepkg",
		Args: []string{
			"--force",
			"--nodeps",
			"--skipchecksums",
			"--packagelist",
		},
		Dir: workdir,
	}); err != nil {
		return "", err
	}

	if err := common.CommandStream(ctx, common.Cmd{
		Name: "makepkg",
		Args: []string{
			"--force",
			"--nodeps",
			"--skipchecksums",
			"--holdver",
		},
		Dir: workdir,
	}); err != nil {
		return "", err
	}

	matches, err := filepath.Glob(filepath.Join(workdir, fmt.Sprintf("%s-%s-1-%s.pkg.tar.*", packageName, pkgVersion, arch)))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("makepkg did not produce a package for %s", arch)
	}

	pkg := filepath.Join(outputDir, filepath.Base(matches[0]))
	if err := os.Rename(matches[0], pkg); err != nil {
		return "", err
	}
	return pkg, nil
}

func sanitizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.ReplaceAll(v, "-", "_")
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0_" + v
	}
	return v
}

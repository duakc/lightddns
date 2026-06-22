// Package deb builds .deb packages. Go only prepares the preconditions and data
// for the real tool: it lays down the packaged files (declared in
// packingFileList, with per-file placeholder substitution from subSet) into a
// staging tree, compiles each shipped target straight into it, then shells out
// to dpkg-deb - the same "orchestrate in Go, run the real tool" idea as build.
package deb

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/build"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"
)

const binaryBase = constpkg.Project

// placeholder keys (rendered as __KEY__ in the source files, see packing).
const (
	subVersion   = "VERSION"
	subArch      = "ARCH"
	subSchemaURL = "SCHEMA_URL"
)

var (
	releaseDirFileHelper = mt.Must(filehelper.New(common.ReleaseDir()))
	debFileHelper        = mt.Must(filehelper.New(common.ReleaseDir("deb", "pkgroot")))
)

// packingFileList enumerates every static file shipped in the .deb. The binary
// is per-arch and compiled separately. DEBIAN/* and etc/* come from the pkgroot
// skeleton; the systemd units and the (gzipped) man page come from release/ and
// land under lib/systemd/system and usr/share/man respectively.
var packingFileList = []packing.File{
	{FS: debFileHelper, From: "DEBIAN/control", To: "DEBIAN/control", Mode: 0o644, SubSetVec: packing.SubSetVec{subVersion, subArch}},
	{FS: debFileHelper, From: "DEBIAN/conffiles", To: "DEBIAN/conffiles", Mode: 0o644},
	{FS: debFileHelper, From: "DEBIAN/postinst", To: "DEBIAN/postinst", Mode: 0o755},
	{FS: debFileHelper, From: "DEBIAN/prerm", To: "DEBIAN/prerm", Mode: 0o755},
	{FS: debFileHelper, From: "DEBIAN/postrm", To: "DEBIAN/postrm", Mode: 0o755},
	{FS: debFileHelper, From: "etc/lightddns.yaml", To: "etc/lightddns.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURL}},
	{FS: debFileHelper, From: "etc/lightddns.d/example.yaml", To: "etc/lightddns.d/example.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURL}},
	{FS: debFileHelper, From: "etc/default/lightddns", To: "etc/default/lightddns", Mode: 0o640},
	{FS: releaseDirFileHelper, From: "systemd/lightddns.service", To: "lib/systemd/system/lightddns.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "systemd/lightddns@.service", To: "lib/systemd/system/lightddns@.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "man/lightddns.1", To: "usr/share/man/man1/lightddns.1.gz", Mode: 0o644, Gzip: true},
}

var (
	outputDir = common.BuildDir("deb") // finished .debs

	buildAll     bool
	buildVersion string
	buildBranch  string

	verbose bool

	pkgVersion string // Debian-valid version (sanitized), used for the package
	schemaURL  string
)

func Run(ctx context.Context) {
	flag.StringVar(&buildVersion, "version", "", "package version (default: git tag or short hash)")
	flag.StringVar(&buildBranch, "branch", "", "build branch (default: current git branch)")
	flag.BoolVar(&buildAll, "all", false, "build every shipped arch (default: host arch only)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()
	common.Verbose = verbose

	// buildVersion is the tag or short hash: it stamps the binary. pkgVersion is
	// its Debian-valid form, for the package. The schema URL is pinned to the
	// build branch so an installed config validates against the matching schema.
	if buildVersion == "" {
		buildVersion = gitver.Version(ctx)
	}

	if buildBranch == "" {
		buildBranch = gitver.Branch(ctx)
	}

	pkgVersion = sanitizeVersion(buildVersion)
	schemaURL = fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/release/schema.json",
		constpkg.Repo, buildVersion, buildBranch)

	built := 0

	for _, tgt := range target.All() {
		if !tgt.Deb || (runtime.GOARCH != tgt.GOARCH || runtime.GOOS != tgt.GOOS) {
			continue
		}

		common.Infof("GOOS=%s GOARCH=%s TARGET_GOOS=%s TARGET_GOARCH=%s",
			runtime.GOOS, runtime.GOARCH, tgt.GOOS, tgt.GOARCH)

		deb, err := pack(ctx, tgt)
		if err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", deb)
		built++
	}

	if built == 0 {
		common.Fatalf("no deb built with GOARCH=%s GOOS=%s", runtime.GOARCH, runtime.GOOS)
	}
	common.Infof("done, built %d package(s)", built)
}

// pack stages one arch under build/draft and runs dpkg-deb, writing the
// finished .deb into the products dir.
func pack(ctx context.Context, tgt target.Target) (string, error) {
	debArch := tgt.DEBArchName()
	pkgName := fmt.Sprintf("%s_%s_%s", binaryBase, pkgVersion, debArch)

	stage := common.BuildDraftDir("deb", pkgName)
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

	subSet := packing.SubSet{
		subVersion:   pkgVersion,
		subArch:      debArch,
		subSchemaURL: schemaURL,
	}

	// static files: substitute placeholders (and gzip where asked) and write
	// into the stage tree. Indexed so the File's lazy-read state isn't copied.
	for i := range packingFileList {
		pf := &packingFileList[i]
		if err := pf.Process(stageFileHelper, subSet); err != nil {
			return "", fmt.Errorf("%s: %w", pf.To, err)
		}
	}

	// Compile straight into the package tree (build.Binary's OutputDir),
	// reusing the shared params with the plain (non-qualified) binary name.
	if err := stageFileHelper.MkdirAll("usr/bin", 0o755); err != nil {
		return "", err
	}
	p := build.DefaultParams()
	p.OutputDir = filepath.Join(stage, "usr/bin")
	p.Qualified = false
	p.BinaryName = binaryBase
	p.Version = buildVersion
	p.Branch = buildBranch

	if _, err := build.Binary(ctx, tgt, p); err != nil {
		return "", err
	}

	deb := filepath.Join(outputDir, pkgName+".deb")
	if err := common.CommandStream(ctx, common.Cmd{
		Name: "dpkg-deb",
		Args: []string{"--root-owner-group", "--build", stage, deb},
	}); err != nil {
		return "", err
	}
	return deb, nil
}

// sanitizeVersion strips a leading "v" and keeps it Debian-valid (digit first).
func sanitizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0~" + v
	}
	return v
}

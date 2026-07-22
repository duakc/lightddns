// Package rpm builds .rpm packages. Go only prepares the preconditions and data
// for the real tool: it lays down the packaged files (declared in
// packingFileList, with per-file placeholder substitution from subSet) into a
// staging tree, compiles each shipped target straight into it, renders the
// .spec, then shells out to rpmbuild - the same "orchestrate in Go, run the
// real tool" idea as build.
//
// rpmbuild wants a .spec and its own buildroot, so the staged tree is the
// SOURCE: the spec's %install copies %{staging} into the buildroot. See
// release/rpm/pkgroot.
package rpm

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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

const binaryBase = constpkg.Project

// placeholder keys (rendered as __KEY__ in the source files, see packing).
const (
	subVersion   = "VERSION"
	subArch      = "ARCH"
	subSchemaURL = "SCHEMA_URL"
)

// pkgRelease is the rpm Release field; we pin "dist" to empty (see pack) so the
// finished filename is deterministic across whichever host runs rpmbuild.
const pkgRelease = "1"

// specName is the source spec inside the rpm pkgroot.
const specName = "lightddns.spec"

var (
	releaseDirFileHelper = mt.Must(filehelper.New(common.ReleaseDir()))
	rpmFileHelper        = mt.Must(filehelper.New(common.ReleaseDir("rpm", "pkgroot")))
)

// packingFileList enumerates every static file shipped in the .rpm. The binary
// is per-arch and compiled separately; the spec is rendered separately (it is
// consumed by rpmbuild, not installed). The example configs, systemd units and
// (gzipped) man page are format-agnostic and come from the shared release/
// subdirs - the single example/lightddns.yaml serves both /etc/lightddns.yaml
// and the .d/ sample. They land under usr/lib/systemd/system and usr/share/man.
var packingFileList = []packing.File{
	{FS: releaseDirFileHelper, From: "example/lightddns.yaml", To: "etc/lightddns.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURL}},
	{FS: releaseDirFileHelper, From: "example/lightddns.yaml", To: "etc/lightddns.d/example.yaml", Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURL}},
	{FS: releaseDirFileHelper, From: "example/environment", To: "etc/default/lightddns", Mode: 0o640},
	{FS: releaseDirFileHelper, From: "systemd/lightddns.service", To: "usr/lib/systemd/system/lightddns.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "systemd/lightddns@.service", To: "usr/lib/systemd/system/lightddns@.service", Mode: 0o644},
	{FS: releaseDirFileHelper, From: "man/lightddns.1", To: "usr/share/man/man1/lightddns.1.gz", Mode: 0o644, Gzip: true},
}

var (
	outputDir = common.BuildDir("rpm") // finished .rpms

	buildAll     bool
	buildVersion string
	buildBranch  string

	verbose bool

	pkgVersion string // rpm-valid version (sanitized), used for the package
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
	// its rpm-valid form, for the package. The schema URL is pinned to the build
	// version so an installed config validates against the matching schema.
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

	targets := target.RPMTargets(buildAll)
	if len(targets) == 0 {
		common.Fatalf("no rpm built with GOARCH=%s GOOS=%s", runtime.GOARCH, runtime.GOOS)
	}

	for _, tgt := range targets {
		common.Infof("GOOS=%s GOARCH=%s TARGET_GOOS=%s TARGET_GOARCH=%s",
			runtime.GOOS, runtime.GOARCH, tgt.GOOS, tgt.GOARCH)

		rpm, err := pack(ctx, tgt)
		if err != nil {
			common.Fatalf("package: %s", err)
		}
		common.Infof("built %s", rpm)
	}
	common.Infof("done, built %d package(s)", len(targets))
}

// pack stages one arch's content tree under build/draft, renders the spec, and
// runs rpmbuild, writing the finished .rpm into the products dir.
func pack(ctx context.Context, tgt target.Target) (string, error) {
	rpmArch := tgt.RPMArchName()
	pkgName := fmt.Sprintf("%s-%s-%s.%s", binaryBase, pkgVersion, pkgRelease, rpmArch)

	// staging is the content tree (becomes the buildroot via the spec's %install).
	stage := common.BuildDraftDir("rpm", "pkgroot-"+rpmArch)
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
		subArch:      rpmArch,
		subSchemaURL: schemaURL,
	}

	// Lay down the static files, then compile the binary straight into the
	// staged tree under usr/bin.
	if err := packing.ProcessAll(stageFileHelper, packingFileList, subSet); err != nil {
		return "", err
	}
	if _, err := gobuild.Plain(ctx, tgt, filepath.Join(stage, "usr/bin"), buildVersion, buildBranch); err != nil {
		return "", err
	}

	// Render the spec (placeholder substitution only) into the draft SPECS dir.
	specPath, err := renderSpec(rpmArch)
	if err != nil {
		return "", err
	}

	// rpmbuild wants absolute paths for its topdir/output/buildroot source.
	stageAbs, err := filepath.Abs(stage)
	if err != nil {
		return "", err
	}
	topdir, err := filepath.Abs(common.BuildDraftDir("rpm", "topdir"))
	if err != nil {
		return "", err
	}
	rpmdir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", err
	}

	rpm := filepath.Join(outputDir, pkgName+".rpm")
	if err := common.CommandStream(ctx, common.Cmd{
		Name: "rpmbuild",
		Args: []string{
			"-bb",
			// build for the target arch even on a foreign-arch host; without
			// this rpmbuild rejects a cross arch ("no compatible architectures").
			"--target", rpmArch,
			"--define", "_topdir " + topdir,
			"--define", "_rpmdir " + rpmdir,
			// flatten the output: no per-arch subdir, name it ourselves.
			"--define", "_rpmfilename " + pkgName + ".rpm",
			// pin Release so the filename is deterministic off any host.
			"--define", "dist %{nil}",
			// the staged content tree the spec's %install copies in.
			"--define", "staging " + stageAbs,
			specPath,
		},
	}); err != nil {
		return "", err
	}
	return rpm, nil
}

// renderSpec substitutes __VERSION__ into the source spec and writes the result
// under the draft SPECS dir, returning its path. The arch is not baked into the
// spec; it is set on the rpmbuild command line via --target.
func renderSpec(rpmArch string) (string, error) {
	raw := rpmFileHelper.MustReadFile(specName)
	rendered, err := packing.SubSet{
		subVersion: pkgVersion,
	}.Process(packing.SubSetVec{subVersion}, slices.Clone(raw))
	if err != nil {
		return "", err
	}

	specDir := common.BuildDraftDir("rpm", "SPECS")
	specPath := filepath.Join(specDir, fmt.Sprintf("%s-%s.spec", binaryBase, rpmArch))
	if err := os.WriteFile(specPath, rendered, 0o644); err != nil {
		return "", err
	}
	return specPath, nil
}

// sanitizeVersion strips a leading "v" and keeps it rpm-valid: rpm forbids "-"
// in the Version field, so dashes (e.g. a "1.2.3-rc1" tag) become "~", which
// rpm orders as a pre-release. A non-numeric lead gets a "0.0.0~" prefix.
func sanitizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.ReplaceAll(v, "-", "~")
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0~" + v
	}
	return v
}

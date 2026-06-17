// Package deb builds .deb packages. Go only prepares the preconditions and data
// for the real tool: it compiles each shipped target (build package) straight
// into a staging tree (the skeleton from release/deb/pkgroot plus units / man),
// fills in the dynamic bits (version, arch, schema URL), then shells out to
// dpkg-deb - the same "orchestrate in Go, run the real tool" idea as build.
package deb

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/build"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
)

const binaryBase = constant.Project

var (
	outputDir  = common.BuildDir("deb") // finished .debs
	pkgrootDir = common.ReleaseDir("deb", "pkgroot")
	systemdDir = common.ReleaseDir("systemd")
	manFile    = common.ReleaseDir("man", "lightddns.1")

	params   = build.DefaultParams() // flags bind onto this; reused for each build
	buildAll bool
	verbose  bool

	pkgVersion string // Debian-valid version (sanitized), used for the package
	schemaURL  string
)

func Run(ctx context.Context) {
	flag.StringVar(&params.Version, "version", "", "package version (default: git tag or short hash)")
	flag.StringVar(&params.Branch, "branch", "", "build branch (default: current git branch)")
	flag.BoolVar(&buildAll, "all", false, "build every shipped arch (default: host arch only)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()
	common.Verbose = verbose

	// params.Version is the tag or short hash: it stamps the binary and is the
	// schema URL ref. pkgVersion is its Debian-valid form, for the package.
	if params.Version == "" {
		params.Version = gitver.Version(ctx)
	}
	if params.Branch == "" {
		params.Branch = gitver.Branch(ctx)
	}
	pkgVersion = sanitizeVersion(params.Version)
	schemaURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/release/schema.json", constant.Repo, params.Version)

	built := 0
	for _, tgt := range target.All() {
		if !tgt.Deb ||
			(!buildAll && runtime.GOARCH != tgt.GOARCH && runtime.GOOS != tgt.GOOS) {
			continue
		}

		debArch, ok := tgt.PackageArch(target.FormatDeb)
		if !ok {
			continue
		}

		deb, err := pack(ctx, tgt, debArch)
		if err != nil {
			common.Fatalf("package %s: %s", debArch, err)
		}
		common.Infof("built %s", deb)
		built++
	}

	if built == 0 {
		common.Fatalf("no deb built with GOARCH=%s GOOS=%s", runtime.GOARCH, runtime.GOOS)
	}
	common.Infof("done: built %d package(s)", built)
}

// pack stages one arch under build/draft and runs dpkg-deb, writing the
// finished .deb into the products dir.
func pack(ctx context.Context, tgt target.Target, debArch string) (string, error) {
	pkgName := fmt.Sprintf("%s_%s_%s", binaryBase, pkgVersion, debArch)
	stage := common.BuildDraftDir("deb", pkgName)
	if err := os.RemoveAll(stage); err != nil {
		return "", err
	}

	// skeleton (DEBIAN/ + etc/), then overlay the per-build files.
	if err := os.CopyFS(stage, os.DirFS(pkgrootDir)); err != nil {
		return "", fmt.Errorf("copy skeleton: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(stage, "etc/lightddns.d"), 0o755); err != nil {
		return "", err
	}

	// Compile straight into the package tree (build.Binary's OutputDir),
	// reusing the shared params with the plain (non-qualified) binary name.
	p := params
	p.OutputDir = filepath.Join(stage, "usr/bin")
	if _, err := build.Binary(ctx, tgt, p); err != nil {
		return "", err
	}

	for _, unit := range []string{"lightddns.service", "lightddns@.service"} {
		if err := put(filepath.Join(systemdDir, unit), filepath.Join(stage, "lib/systemd/system", unit), 0o644); err != nil {
			return "", err
		}
	}
	if man, err := os.ReadFile(manFile); err == nil {
		if err := write(filepath.Join(stage, "usr/share/man/man1/lightddns.1.gz"), gzipBytes(man), 0o644); err != nil {
			return "", err
		}
	} else {
		common.Warnf("no man page at %s (skipping)", manFile)
	}

	// dynamic data + permissions (config/secrets not world-readable).
	if err := subst(filepath.Join(stage, "DEBIAN/control"), "__VERSION__", pkgVersion, "__ARCH__", debArch); err != nil {
		return "", err
	}
	if err := subst(filepath.Join(stage, "etc/lightddns.yaml"), "__SCHEMA_URL__", schemaURL); err != nil {
		return "", err
	}
	for path, mode := range map[string]os.FileMode{
		"DEBIAN/postinst":       0o755,
		"DEBIAN/prerm":          0o755,
		"DEBIAN/postrm":         0o755,
		"etc/lightddns.yaml":    0o640,
		"etc/default/lightddns": 0o640,
	} {
		if err := os.Chmod(filepath.Join(stage, path), mode); err != nil {
			return "", err
		}
	}

	deb := filepath.Join(outputDir, pkgName+".deb")
	if err := common.Stream(ctx, common.Cmd{
		Name: "dpkg-deb",
		Args: []string{"--root-owner-group", "--build", stage, deb},
	}); err != nil {
		return "", err
	}
	return deb, nil
}

// put copies src to dst with the given mode; write writes bytes to dst.
func put(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return write(dst, data, mode)
}

func write(dst string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}

// subst replaces old/new pairs in a file in place.
func subst(path string, pairs ...string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.NewReplacer(pairs...).Replace(string(data))), 0o644)
}

func gzipBytes(data []byte) []byte {
	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, _ = gz.Write(data)
	_ = gz.Close()
	return buf.Bytes()
}

// sanitizeVersion strips a leading "v" and keeps it Debian-valid (digit first).
func sanitizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v[0] < '0' || v[0] > '9' {
		v = "0.0.0~" + v
	}
	return v
}

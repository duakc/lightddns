package build

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/buildprofile"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
)

const defaultWorkdir = "./cmd/lightddns/"

type Params struct {
	Version string
	Branch  string

	WorkingDir string // package to build (go build target dir)
	OutputDir  string // directory the binary is written into
	BinaryName string // base binary name
	ExtraTags  []string

	// Qualified writes the platform-qualified name (lightddns-linux-amd64);
	// otherwise the plain BinaryName is used (single-target build).
	Qualified bool
}

func DefaultParams() Params {
	return Params{
		WorkingDir: defaultWorkdir,
		OutputDir:  common.BuildDir("bin"),
		BinaryName: constant.Project,
		Qualified:  true,
	}
}

func Binary(ctx context.Context, tgt target.Target, p Params) (string, error) {
	const main = "."
	goos, goarch := tgt.GOOS, tgt.GOARCH

	// ldflags: target defaults + version/branch stamps.
	ldflags := append([]string{}, tgt.LDFlags...)
	ldflags = append(ldflags,
		fmt.Sprintf(`-X "github.com/%s/constant.Version=%s"`, constant.Repo, p.Version),
		fmt.Sprintf(`-X "github.com/%s/constant.Branch=%s"`, constant.Repo, p.Branch))

	tags := append([]string{}, tgt.TAGS...)
	tags = append(tags, p.ExtraTags...)
	debug := tgt.Debug || slices.Contains(p.ExtraTags, "debug")

	// environment
	env := []string{"GOOS=" + goos, "GOARCH=" + goarch}
	cgo := "0"
	if tgt.CGO {
		cgo = "1"
	}
	env = append(env, "CGO_ENABLED="+cgo)
	if tgt.GOAMD64Version != 0 && goarch == "amd64" {
		env = append(env, fmt.Sprintf("GOAMD64=v%d", tgt.GOAMD64Version))
	}
	if tgt.SoftFloat && goarch == "mips" {
		env = append(env, "GOMIPS=softfloat")
	}
	if tgt.GOARMVersion > 0 && goarch == "arm" {
		env = append(env, fmt.Sprintf("GOARM=%d", tgt.GOARMVersion))
	}

	args := []string{"build", "-C", main}

	// platform-qualified name; debug/.exe are build-mode suffixes.
	name := tgt.BinaryName(p.BinaryName)
	if debug {
		name += "-debug"
		if !slices.Contains(tags, "debug") {
			tags = append(tags, "debug")
		}
	} else {
		ldflags = append(ldflags, "-w", "-s")
		args = append(args, "-trimpath")
	}

	if len(tags) > 0 {
		joined := strings.Join(tags, ",")
		args = append(args, "-tags", joined)
		// Stamp the build tags so `version` can report what was compiled in.
		ldflags = append(ldflags,
			fmt.Sprintf(`-X "github.com/%s/constant.Tags=%s"`, constant.Repo, joined))
	}
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}

	if !p.Qualified {
		name = p.BinaryName
	}
	if goos == "windows" {
		name += ".exe"
	}

	outPath := filepath.Join(p.OutputDir, name)
	args = append(args, "-o", outPath, p.WorkingDir)
	if err := common.CommandStream(ctx, common.Cmd{Name: "go", Args: args, Env: env}); err != nil {
		return "", err
	}
	return outPath, nil
}

// Plain builds tgt's binary with the plain (unqualified) name into outputDir,
// creating outputDir if needed. It is the shared "compile straight into a
// staging tree" step used by the package builders.
func Plain(ctx context.Context, tgt target.Target, outputDir, version, branch string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	p := DefaultParams()
	p.OutputDir = outputDir
	p.Qualified = false
	p.Version = version
	p.Branch = branch
	return Binary(ctx, tgt, p)
}

func Run(ctx context.Context) {
	params := DefaultParams()
	var (
		tags    string
		verbose bool
		all     bool
		goos    string
		goarch  string
	)
	// Per-invocation knobs stay as flags/env; the build profile holds only the
	// common defaults (tags, output name). The target arch is chosen here:
	// --os and --arch each narrow the matrix independently, so --os linux builds
	// every linux arch, --arch arm64 builds arm64 across every OS, and both
	// together pin a single target. Neither (and no --all) builds the host.
	flag.StringVar(&params.Version, "version", "", "version (default: git tag or short hash)")
	flag.StringVar(&params.Branch, "branch", "", "branch (default: current git branch)")
	flag.StringVar(&params.WorkingDir, "workdir", params.WorkingDir, "package to build")
	flag.StringVar(&params.BinaryName, "binary", params.BinaryName, "binary name")
	flag.StringVar(&tags, "tags", "", "extra build tags, appended to the profile's (comma separated)")
	flag.StringVar(&goos, "os", "", "GOOS to build (e.g. linux); empty matches every OS")
	flag.StringVar(&goarch, "arch", "", "GOARCH to build (e.g. amd64); empty matches every arch")
	flag.BoolVar(&all, "all", false, "build every target (all OS/arch)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	common.Verbose = verbose
	if params.Version == "" {
		params.Version = gitver.Version(ctx)
	}
	if params.Branch == "" {
		params.Branch = gitver.Branch(ctx)
	}

	var flagTags []string
	if tags != "" {
		flagTags = strings.Split(tags, ",")
	}

	targets := resolveTargets(all, goos, goarch)
	if len(targets) == 0 {
		common.Fatalf("no target matches --os %q --arch %q (host GOOS=%s GOARCH=%s)",
			goos, goarch, runtime.GOOS, runtime.GOARCH)
	}
	params.Qualified = len(targets) > 1

	profiles, err := buildprofile.Load()
	if err != nil {
		common.Fatalf("load build profile: %s", err)
	}
	enabled := buildprofile.Enabled(profiles)

	// No profile (missing file or none enabled): build with the flag tags only.
	if len(enabled) == 0 {
		params.ExtraTags = flagTags
		buildTargets(ctx, targets, params)
		return
	}

	for _, prof := range enabled {
		p := params
		p.ExtraTags = append(append([]string{}, prof.DefaultTags...), flagTags...)
		p.OutputDir = binOutputDir(prof.BuildName)
		buildTargets(ctx, targets, p)
	}
}

// buildTargets compiles every target with the given params.
func buildTargets(ctx context.Context, targets []target.Target, p Params) {
	for _, tgt := range targets {
		if _, err := Binary(ctx, tgt, p); err != nil {
			common.Fatalf("%s", err.Error())
		}
	}
}

// resolveTargets picks the targets to build. An empty goos/goarch is a wildcard
// for that field; with --all (or both empty via no host match) it is every
// target. With neither flag set it is just the build host's arch.
func resolveTargets(all bool, goos, goarch string) []target.Target {
	if !all && goos == "" && goarch == "" {
		return hostTargets()
	}
	var out []target.Target
	for _, t := range target.All() {
		if (goos == "" || t.GOOS == goos) && (goarch == "" || t.GOARCH == goarch) {
			out = append(out, t)
		}
	}
	return out
}

// hostTargets is the single target matching the build host's GOOS/GOARCH.
func hostTargets() []target.Target {
	for _, t := range target.All() {
		if t.GOOS == runtime.GOOS && t.GOARCH == runtime.GOARCH {
			return []target.Target{t}
		}
	}
	return nil
}

// binOutputDir is build/bin, or build/build_<buildName>/bin when named.
func binOutputDir(buildName string) string {
	if buildName == "" {
		return common.BuildDir("bin")
	}
	return common.BuildDir("build_"+buildName, "bin")
}

package build

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/duakc/lightddns/constant"
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
		args = append(args, "-tags", strings.Join(tags, ","))
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

func Run(ctx context.Context) {
	params := DefaultParams()
	var (
		tags    string
		verbose bool
	)
	flag.StringVar(&params.Version, "version", "", "version (default: git tag or short hash)")
	flag.StringVar(&params.Branch, "branch", "", "branch (default: current git branch)")
	flag.StringVar(&params.WorkingDir, "workdir", params.WorkingDir, "package to build")
	flag.StringVar(&params.BinaryName, "binary", params.BinaryName, "binary name")
	flag.StringVar(&tags, "tags", "", "extra build tags (comma separated)")
	flag.BoolVar(&params.Qualified, "all", false, "build every target (qualified names)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	common.Verbose = verbose
	if params.Version == "" {
		params.Version = gitver.Version(ctx)
	}
	if params.Branch == "" {
		params.Branch = gitver.Branch(ctx)
	}
	if tags != "" {
		params.ExtraTags = strings.Split(tags, ",")
	}

	for _, tgt := range target.All() {
		if !params.Qualified && !(runtime.GOARCH == tgt.GOARCH && runtime.GOOS == tgt.GOOS) {
			continue
		}
		if _, err := Binary(ctx, tgt, params); err != nil {
			common.Fatalf("%s", err.Error())
		}
	}
}

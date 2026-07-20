package gobuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	goyaml "github.com/goccy/go-yaml"
)

const EnvBuildProfilePath = "RELEASE_BUILD_PROFILE"

type Params struct {
	Version string `yaml:"version"`
	Branch  string `yaml:"branch"`

	WorkingDir string `yaml:"workingDir"` // package to build (go build target dir)
	OutputDir  string `yaml:"outputDir"`  // directory the binary is written into
	BinaryName string `yaml:"binaryName"` // base binary name

	ExtraTags []string          `yaml:"tags"`
	ExtraEnv  map[string]string `yaml:"env"`

	// BuildName names the output tree for the outer build command: empty ->
	// build/bin, else build/build_<buildName>/bin.
	BuildName string `yaml:"buildName"`

	// Qualified writes the platform-qualified name (lightddns-linux-amd64);
	// otherwise the plain BinaryName is used (single-target build).
	Qualified bool `yaml:"qualified"`
}

func DefaultParams() Params {
	p, err := loadBuildProfile()
	if err != nil {
		common.Fatalf("load build profile: %s", err)
	}
	if p.BuildName != "" {
		p.OutputDir = common.BuildDir("build_"+p.BuildName, "bin")
	}
	return p
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

	env := []string{"GOOS=" + goos, "GOARCH=" + goarch}
	cgo := "0"
	if tgt.CGO {
		cgo = "1"
	}
	env = append(env, "CGO_ENABLED="+cgo)
	for k, v := range p.ExtraEnv {
		env = append(env, k+"="+v)
	}
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

func loadBuildProfile() (Params, error) {
	path := buildProfilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Params{}, err
	}
	var p Params
	if err := goyaml.Unmarshal(data, &p); err != nil {
		return Params{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return p, nil
}

func buildProfilePath() string {
	if p := os.Getenv(EnvBuildProfilePath); p != "" {
		return p
	}
	return common.ReleaseDir("build.yaml")
}

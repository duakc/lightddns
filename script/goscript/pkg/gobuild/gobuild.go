package gobuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/buildinfo"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/target"
)

type Params struct {
	WorkingDir string // package to build (go build target dir)
	OutputDir  string // directory the binary is written into
	BinaryName string // base binary name

	ExtraTags []string
	ExtraEnv  []string
	ExtraArgs []string

	LDFlags []string

	// Qualified writes the platform-qualified name (lightddns-linux-amd64);
	// otherwise the plain BinaryName is used (single-target build).
	Qualified bool
}

func Binary(ctx context.Context, tgt target.Target, p Params) (string, error) {
	const main = "."
	goos, goarch := tgt.GOOS, tgt.GOARCH
	if err := p.validate(); err != nil {
		return "", err
	}

	// ldflags: target defaults plus the caller-supplied build flags.
	ldflags := append([]string{}, tgt.LDFlags...)
	ldflags = append(ldflags, p.LDFlags...)

	tags := append([]string{}, tgt.TAGS...)
	tags = append(tags, p.ExtraTags...)
	debug := tgt.Debug || slices.Contains(p.ExtraTags, "debug")

	env := []string{"GOOS=" + goos, "GOARCH=" + goarch}
	cgo := "0"
	if tgt.CGO {
		cgo = "1"
	}
	env = append(env, "CGO_ENABLED="+cgo)
	if tgt.GOAMD64Version != 0 && goarch == "amd64" {
		env = append(env, fmt.Sprintf("GOAMD64=v%d", tgt.GOAMD64Version))
	}
	if tgt.MIPSSoftFloat && (goarch == "mips" || goarch == "mipsle") {
		env = append(env, "GOMIPS=softfloat")
	}
	if tgt.MIPSSoftFloat && (goarch == "mips64" || goarch == "mips64le") {
		env = append(env, "GOMIPS64=softfloat")
	}
	if tgt.GOARMVersion > 0 && goarch == "arm" {
		env = append(env, fmt.Sprintf("GOARM=%d", tgt.GOARMVersion))
	}
	if tgt.GO386 != "" && goarch == "386" {
		env = append(env, "GO386="+tgt.GO386)
	}

	args := []string{"build", "-C", main}

	name := binaryName(tgt, p, buildinfo.Version())
	if debug {
		name += "-debug"
		if !slices.Contains(tags, "debug") {
			tags = append(tags, "debug")
		}
	} else {
		ldflags = append(ldflags, "-w", "-s")
		args = append(args, "-trimpath")
	}

	joinedTags := strings.Join(tags, ",")
	if len(tags) > 0 {
		args = append(args, "-tags", joinedTags)
	}

	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}

	outPath := filepath.Join(p.OutputDir, name)
	if err := os.MkdirAll(p.OutputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory %s: %w", p.OutputDir, err)
	}

	env = append(env, p.ExtraEnv...)
	args = append(args, p.ExtraArgs...)
	args = append(args, "-o", outPath, p.WorkingDir)

	if err := common.CommandStream(ctx, common.Cmd{Name: "go", Args: args, Env: env}); err != nil {
		return "", err
	}
	return outPath, nil
}

func binaryName(tgt target.Target, p Params, version string) string {
	name := p.BinaryName
	if p.Qualified {
		name = tgt.BinaryName(name)
	}
	if version = versionSuffix(version); version != "" {
		name += version
	}
	if tgt.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func versionSuffix(version string) string {
	version = strings.TrimSpace(version)
	if version == "" || version == "(unknown)" {
		return ""
	}
	version = strings.NewReplacer(
		"/", "-",
		"\\", "-",
		" ", "-",
	).Replace(version)
	return "-" + version
}

func (p Params) validate() error {
	if p.WorkingDir == "" {
		return fmt.Errorf("missing --%s", buildParamFlagName("workdir"))
	}
	if p.OutputDir == "" {
		return fmt.Errorf("missing --%s", buildParamFlagName("output"))
	}
	if p.BinaryName == "" {
		return fmt.Errorf("missing --%s", buildParamFlagName("binary_name"))
	}
	return nil
}

package gobuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gitver"
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

	ExtraTags []string `yaml:"tags"`
	ExtraEnv  []string `yaml:"env"`

	// LDFlags are the raw ldflag templates from the build profile. They may use
	// ${VAR} placeholders (see buildVarExpander) so the linker symbol paths stay
	// in the profile, not hard-coded in this builder.
	LDFlags []string `yaml:"ldflags"`

	// Qualified writes the platform-qualified name (lightddns-linux-amd64);
	// otherwise the plain BinaryName is used (single-target build).
	Qualified bool `yaml:"qualified"`
}

func DefaultParams() Params {
	p, err := loadBuildProfile()
	if err != nil {
		common.Fatalf("load build profile: %s", err)
	}
	return p
}

func Binary(ctx context.Context, tgt target.Target, p Params) (string, error) {
	const main = "."
	goos, goarch := tgt.GOOS, tgt.GOARCH

	// ldflags: target defaults + the stamps configured in the build profile.
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

	joinedTags := strings.Join(tags, ",")
	if len(tags) > 0 {
		args = append(args, "-tags", joinedTags)
	}

	// Expand ${VAR} placeholders so the profile owns the symbol paths.
	expand := buildVarExpander(p, tgt, joinedTags)
	for i := range ldflags {
		ldflags[i] = os.Expand(ldflags[i], expand)
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

	env = append(env, p.ExtraEnv...)
	args = append(args, "-o", outPath, p.WorkingDir)

	if err := common.CommandStream(ctx, common.Cmd{Name: "go", Args: args, Env: env}); err != nil {
		return "", err
	}
	return outPath, nil
}

func Plain(ctx context.Context, tgt target.Target, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	p := DefaultParams()
	p.OutputDir = outputDir
	p.Qualified = false
	p.Version = gitver.Version(ctx)
	p.Branch = gitver.Branch(ctx)
	return Binary(ctx, tgt, p)
}

// buildVarExpander resolves ${VAR} placeholders in the profile ldflags from the
// exported fields of Params and Target (Version -> PARAM_BUILD_VERSION, GOARCH ->
// TARGET_GOARCH), plus REPO_NAME and the effective PARAM_BUILD_TAGS. Unknown keys
// fall back to the environment.
func buildVarExpander(p Params, tgt target.Target, tags string) func(string) string {
	vars := map[string]string{
		"REPO_NAME":        constant.Repo,
		"PARAM_BUILD_TAGS": tags,
	}
	for _, src := range []struct {
		prefix string
		val    reflect.Value
	}{
		{"PARAM_BUILD", reflect.ValueOf(p)},
		{"TARGET", reflect.ValueOf(tgt)},
	} {
		t := src.val.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			var s string
			switch fv := src.val.Field(i); fv.Kind() {
			case reflect.String:
				s = fv.String()
			case reflect.Bool:
				s = strconv.FormatBool(fv.Bool())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				s = strconv.FormatInt(fv.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				s = strconv.FormatUint(fv.Uint(), 10)
			case reflect.Slice:
				if fv.Type().Elem().Kind() != reflect.String {
					continue
				}
				parts := make([]string, fv.Len())
				for j := range parts {
					parts[j] = fv.Index(j).String()
				}
				s = strings.Join(parts, ",")
			default:
				continue
			}
			vars[src.prefix+"_"+screamingSnake(f.Name)] = s
		}
	}
	return func(key string) string {
		if v, ok := vars[key]; ok {
			return v
		}
		return os.Getenv(key)
	}
}

// screamingSnake converts a Go field name to SCREAMING_SNAKE_CASE, keeping
// acronyms intact: WorkingDir -> WORKING_DIR, GOAMD64Version -> GOAMD64_VERSION.
func screamingSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || nextLower {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
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

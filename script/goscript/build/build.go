package build

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
)

type BuildTarget struct {
	CGO   bool
	Debug bool

	GOOS    string
	GOARCH  string
	TAGS    []string
	LDFlags []string

	// platform specified
	// amd64
	GOAMD64Version int

	// arm
	GOARMVersion int

	// mips
	SoftFloat bool
}

var (
	version string
	branch  string

	binaryName string

	workingDir string
	outputDir  string
	verbose    bool

	buildAll bool

	extraTags string
)

var allTarget []BuildTarget

func init() {
	// darwin
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
	}...)

	// android
	allTarget = append(allTarget, []BuildTarget{
		//{GOOS: "aix", GOARCH: "ppc64"},
		//{GOOS: "android", GOARCH: "386"},
		//{GOOS: "android", GOARCH: "amd64"},
		//{GOOS: "android", GOARCH: "arm"},
		//{GOOS: "android", GOARCH: "arm64"},
	}...)

	// freebsd
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "freebsd", GOARCH: "386"},
		{GOOS: "freebsd", GOARCH: "amd64"},
		{GOOS: "freebsd", GOARCH: "arm", GOARMVersion: 6},
		{GOOS: "freebsd", GOARCH: "arm", GOARMVersion: 7},
		{GOOS: "freebsd", GOARCH: "arm64"},
		{GOOS: "freebsd", GOARCH: "riscv64"},

		// Note: since freebsd 14 ,all mips architectures are not supported
		//// mips && mipsle
		//{GOOS: "freebsd", GOARCH: "mips", SoftFloat: true},
		//{GOOS: "freebsd", GOARCH: "mipsle", SoftFloat: true},
		//// mipshf && mipslehf
		//{GOOS: "freebsd", GOARCH: "mips", SoftFloat: false},
		//{GOOS: "freebsd", GOARCH: "mipsle", SoftFloat: false},})
	}...)

	// linux-amd64
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 1},
		{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 2},
		{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3},
	}...)

	// linux-arm
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "linux", GOARCH: "arm", GOARMVersion: 5},
		{GOOS: "linux", GOARCH: "arm", GOARMVersion: 6},
		{GOOS: "linux", GOARCH: "arm", GOARMVersion: 7},
		{GOOS: "linux", GOARCH: "arm64"},
	}...)

	// linux-mips
	allTarget = append(allTarget, []BuildTarget{
		// mips && mipsle
		{GOOS: "linux", GOARCH: "mips", SoftFloat: true},
		{GOOS: "linux", GOARCH: "mipsle", SoftFloat: true},
		// mipshf && mipslehf
		{GOOS: "linux", GOARCH: "mips", SoftFloat: false},
		{GOOS: "linux", GOARCH: "mipsle", SoftFloat: false},
	}...)

	// linux-others
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "loong64"},
		//{GOOS: "linux", GOARCH: "ppc64"},
		//{GOOS: "linux", GOARCH: "ppc64le"},
		{GOOS: "linux", GOARCH: "riscv64"},
		//{GOOS: "linux", GOARCH: "s390x"},
		//{GOOS: "linux", GOARCH: "sparc64"},
	}...)

	// openbsd
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "openbsd", GOARCH: "386"},
		{GOOS: "openbsd", GOARCH: "amd64"},
		{GOOS: "openbsd", GOARCH: "arm", GOARMVersion: 7},
		{GOOS: "openbsd", GOARCH: "arm64"},
		{GOOS: "openbsd", GOARCH: "ppc64"},
		{GOOS: "openbsd", GOARCH: "riscv64"},
	}...)

	// windows
	allTarget = append(allTarget, []BuildTarget{
		{GOOS: "windows", GOARCH: "386"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
	}...)
	// others
	allTarget = append(allTarget, []BuildTarget{
		//{GOOS: "aix", GOARCH: "ppc64"},
		//{GOOS: "dragonfly", GOARCH: "amd64"},

		//{GOOS: "illumos", GOARCH: "amd64"},
		//{GOOS: "ios", GOARCH: "amd64"},
		//{GOOS: "ios", GOARCH: "arm64"},
		//{GOOS: "js", GOARCH: "wasm"},

		//{GOOS: "netbsd", GOARCH: "386"},
		//{GOOS: "netbsd", GOARCH: "amd64"},
		//{GOOS: "netbsd", GOARCH: "arm"},
		//{GOOS: "netbsd", GOARCH: "arm64"},

		//{GOOS: "plan9", GOARCH: "386"},
		//{GOOS: "plan9", GOARCH: "amd64"},
		//{GOOS: "plan9", GOARCH: "arm"},
		//{GOOS: "solaris", GOARCH: "amd64"},
		//{GOOS: "wasip1", GOARCH: "wasm"},
	}...)
}

func Run() {
	const Main = "../../"

	flag.StringVar(&version, "version", "0.0.1", "version")
	flag.StringVar(&branch, "branch", "main", "branch")
	flag.StringVar(&workingDir, "workdir", "./cmd/lightddns/", "working directory")
	flag.StringVar(&outputDir, "output", "bin/build", "output directory")
	flag.StringVar(&extraTags, "tags", "", "extra tags")
	flag.StringVar(&binaryName, "binary", "lightddns", "binary name")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.BoolVar(&buildAll, "all", false, "build all")

	flag.Parse()

	if err := os.MkdirAll(filepath.Join(Main, outputDir), 0o755); err != nil {
		fatalErrorf("mkdir: %s", err.Error())
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill,
		syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	defer cancel()

	for _, target := range allTarget {

		goos := target.GOOS
		goarch := target.GOARCH
		if !buildAll && !(runtime.GOARCH == goarch && runtime.GOOS == goos) {
			continue
		}

		// defaults
		target.LDFlags = append(target.LDFlags,
			fmt.Sprintf(`-X "github.com/duakc/lightddns/constant.Version=%s"`, version),
			fmt.Sprintf(`-X "github.com/duakc/lightddns/constant.Branch=%s"`, branch))
		// flags
		if extraTags != "" {
			ext := strings.Split(extraTags, ",")
			if slices.Contains(ext, "debug") {
				target.Debug = true
			}
			target.TAGS = append(target.TAGS, ext...)
		}

		binNameJoin := []string{binaryName, goos, goarch}
		env := append([]string{}, "GOOS="+goos, "GOARCH="+goarch)
		cgo := "0"
		if target.CGO {
			cgo = "1"
		}
		env = append(env, "CGO_ENABLED="+cgo)
		if target.GOAMD64Version != 0 && target.GOARCH == "amd64" {
			binNameJoin = append(binNameJoin, fmt.Sprintf("v%d", target.GOAMD64Version))
			env = append(env, fmt.Sprintf("GOAMD64=v%d", target.GOAMD64Version))
		}
		if target.SoftFloat && target.GOARCH == "mips" {
			binNameJoin = append(binNameJoin, "softfloat")
			env = append(env, "GOMIPS=softfloat")
		} else if target.GOARCH == "mips" || target.GOARCH == "mipsle" {
			binNameJoin = append(binNameJoin, "hardfloat")
		}

		if target.GOARMVersion > 0 && target.GOARCH == "arm" {
			binNameJoin = append(binNameJoin, fmt.Sprintf("v%d", target.GOARMVersion))
			env = append(env, fmt.Sprintf("GOARM=%d", target.GOARMVersion))
		}

		args := []string{"build", "-C", Main}

		if target.Debug {
			binNameJoin = append(binNameJoin, "debug")
			target.TAGS = append(target.TAGS, "debug")
		} else {
			target.LDFlags = append(target.LDFlags, "-w", "-s")
			args = append(args, "-trimpath")
		}

		if len(target.TAGS) > 0 {
			args = append(args, "-tags", strings.Join(target.TAGS, ","))
		}

		if len(target.LDFlags) > 0 {
			args = append(args, "-ldflags", strings.Join(target.LDFlags, " "))
		}

		binName := strings.Join(binNameJoin, "-")
		if goos == "windows" {
			binName += ".exe"
		}
		args = append(args,
			"-o", filepath.Join(outputDir, binName), workingDir)
		if verbose {
			fmt.Printf("$ %s %s\n",
				strings.Join(env, " "),
				strings.Join(append([]string{"go"}, mapQuota(args)...), " "))
		}
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Env = append(os.Environ(), env...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fatalErrorf("%s", err.Error())
		}
	}
}

func fatalErrorf(format string, vv ...any) {
	_, _ = fmt.Fprintf(os.Stderr, ">>>>Fatal: "+format+"\n", vv...)
	os.Exit(1)
}

func mapQuota(s []string) []string {
	v := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		ss := s[i]
		if strings.IndexByte(ss, ' ') >= 0 {
			v[i] = fmt.Sprintf("'%s'", ss)
		} else {
			v[i] = ss
		}
	}
	return v
}

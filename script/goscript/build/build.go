package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	binaryName string

	workingDir string
	outputDir  string
	verbose    bool
)

var allTarget []BuildTarget

func init() {
	flag.StringVar(&workingDir, "wd", ".", "working directory")
	flag.StringVar(&outputDir, "output", ".", "output directory")
	flag.StringVar(&binaryName, "binary", "", "binary name")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

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

func main() {
	flag.Parse()

	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
	}
	outputDir = filepath.Join(workingDir, workingDir)
	if outputDir == "" {
		outputDir = "."
	} else {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			fmt.Printf("mkdir failed: %v\n", err)
			os.Exit(1)
		}
	}

	for _, target := range allTarget {
		goos := target.GOOS
		goarch := target.GOARCH
		ext := ""
		if goos == "windows" {
			ext = ".exe"
		}
		outFile := filepath.Join(outputDir, binaryName+"-"+goos+"-"+goarch+ext)
		args := []string{"build", "-o", outFile}
		if len(target.TAGS) > 0 {
			args = append(args, "-tags="+strings.Join(target.TAGS, ","))
		}
		if len(target.LDFlags) > 0 {
			args = append(args, "-ldflags="+strings.Join(target.LDFlags, " "))
		}
		if target.Debug {
			args = append(args, "-gcflags=all=-N -l")
		}
		env := append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
		cgo := "0"
		if target.CGO {
			cgo = "1"
		}
		env = append(env, "CGO_ENABLED="+cgo)
		if target.GOAMD64Version != 0 {
			env = append(env, fmt.Sprintf("GOAMD64=v%d", target.GOAMD64Version))
		}
		if target.SoftFloat {
			env = append(env, "GOMIPS=softfloat")
		}
		if verbose {
			fmt.Printf("Building for %s/%s -> %s\n", goos, goarch, outFile)
		}
		cmd := exec.Command("go", args...)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Build failed for %s/%s: %v\n", goos, goarch, err)
			os.Exit(1)
		}
	}
}

func verboseMessage(format string, vv ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "verbose: "+format, vv)
}

func fatalError(format string, vv ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "fatal: "+format, vv)
}

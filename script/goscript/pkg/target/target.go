// Package target is the single source of truth for cross-compilation targets:
// the GOOS/GOARCH matrix, how each target's binary is named, and - declared
// right in the table - whether each target ships as a .deb / .rpm.
//
// Both the builder (script/goscript/build) and the packaging scripts consume
// it, so adding an architecture in ONE place flows everywhere, and callers can
// enumerate exactly the targets a given package format ships (on-demand builds).
//
// The Deb/RPM flags are filled in for every row we know about - including the
// commented-out, not-yet-enabled ones - so whoever enables a rare OS/arch later
// does not have to figure out its packaging support: it is already decided here.
package target

import (
	"fmt"
	"strings"
)

// Target describes a single cross-compilation target.
type Target struct {
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

	// Whether this row ships as a .deb / .rpm. Only ONE variant per arch is
	// marked (e.g. baseline amd64, hardfloat mips); microarch/softfloat
	// duplicates stay false. The arch NAME is derived by archName(), not
	// repeated per row.
	Deb bool
	RPM bool
}

// Format identifies an OS packaging system. Architecture naming differs per
// system, which is why the name is translated by archName() per format.
type Format string

const (
	FormatDeb Format = "deb"
	FormatRPM Format = "rpm"
)

// all is the full target matrix. Commented-out entries are kept for the record
// (and pre-tagged with their packaging support for whenever they're enabled).
var all = []Target{
	// darwin (no deb/rpm)
	{GOOS: "darwin", GOARCH: "amd64"},
	{GOOS: "darwin", GOARCH: "arm64"},

	// android (no deb/rpm)
	//{GOOS: "aix", GOARCH: "ppc64"},
	//{GOOS: "android", GOARCH: "386"},
	//{GOOS: "android", GOARCH: "amd64"},
	//{GOOS: "android", GOARCH: "arm"},
	//{GOOS: "android", GOARCH: "arm64"},

	// freebsd (uses pkg, no deb/rpm)
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
	//{GOOS: "freebsd", GOARCH: "mipsle", SoftFloat: false},

	// linux-amd64 (only the baseline ships; microarch levels are extra builds)
	{GOOS: "linux", GOARCH: "amd64", Deb: true, RPM: true},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 1},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 2},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3},

	// linux-arm (rpm only packages the v7/hardfloat profile)
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 5, Deb: true},
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 6, Deb: true},
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 7, Deb: true, RPM: true},
	{GOOS: "linux", GOARCH: "arm64", Deb: true, RPM: true},

	// linux-mips (only hardfloat ships; rpm has no mips)
	// mips && mipsle
	{GOOS: "linux", GOARCH: "mips", SoftFloat: true},
	{GOOS: "linux", GOARCH: "mipsle", SoftFloat: true},
	// mipshf && mipslehf
	{GOOS: "linux", GOARCH: "mips", SoftFloat: false, Deb: true},
	{GOOS: "linux", GOARCH: "mipsle", SoftFloat: false, Deb: true},

	// linux-others
	{GOOS: "linux", GOARCH: "386", Deb: true, RPM: true},
	{GOOS: "linux", GOARCH: "loong64", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "ppc64", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "ppc64le", Deb: true, RPM: true},
	{GOOS: "linux", GOARCH: "riscv64", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "s390x", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "sparc64", Deb: true},

	// openbsd (uses pkg_add, no deb/rpm)
	{GOOS: "openbsd", GOARCH: "386"},
	{GOOS: "openbsd", GOARCH: "amd64"},
	{GOOS: "openbsd", GOARCH: "arm", GOARMVersion: 7},
	{GOOS: "openbsd", GOARCH: "arm64"},
	{GOOS: "openbsd", GOARCH: "ppc64"},
	{GOOS: "openbsd", GOARCH: "riscv64"},

	// windows (no deb/rpm)
	{GOOS: "windows", GOARCH: "386"},
	{GOOS: "windows", GOARCH: "amd64"},
	{GOOS: "windows", GOARCH: "arm64"},

	// others (no deb/rpm)
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
}

func All() []Target {
	return append([]Target(nil), all...)
}

func (t Target) variantParts() []string {
	parts := []string{t.GOOS, t.GOARCH}

	if t.GOAMD64Version != 0 && t.GOARCH == "amd64" {
		parts = append(parts, fmt.Sprintf("v%d", t.GOAMD64Version))
	}

	if t.SoftFloat && t.GOARCH == "mips" {
		parts = append(parts, "softfloat")
	} else if t.GOARCH == "mips" || t.GOARCH == "mipsle" {
		parts = append(parts, "hardfloat")
	}

	if t.GOARMVersion > 0 && t.GOARCH == "arm" {
		parts = append(parts, fmt.Sprintf("v%d", t.GOARMVersion))
	}

	return parts
}

func (t Target) BinaryName(base string) string {
	return strings.Join(append([]string{base}, t.variantParts()...), "-")
}

func (t Target) PackageArch(f Format) (string, bool) {
	switch f {
	case FormatDeb:
		if !t.Deb {
			return "", false
		}
	case FormatRPM:
		if !t.RPM {
			return "", false
		}
	default:
		return "", false
	}
	return archName(t, f), true
}

// archName translates a Go arch to the packaging system's arch name. Only the
// names that differ from GOARCH are listed; everything else maps 1:1.
func archName(t Target, f Format) string {
	switch f {
	case FormatDeb:
		switch t.GOARCH {
		case "arm":
			if t.GOARMVersion == 7 {
				return "armhf"
			}
			return "armel"
		case "386":
			return "i386"
		case "mipsle":
			return "mipsel"
		case "ppc64le":
			return "ppc64el"
		}
	case FormatRPM:
		switch t.GOARCH {
		case "amd64":
			return "x86_64"
		case "arm64":
			return "aarch64"
		case "arm":
			return "armv7hl"
		case "386":
			return "i686"
		case "loong64":
			return "loongarch64"
		}
	}

	// amd64/arm64/mips/riscv64/loong64/ppc64/s390x/sparc64 (deb) and
	// riscv64/ppc64/ppc64le/s390x (rpm) all map straight through.
	return t.GOARCH
}

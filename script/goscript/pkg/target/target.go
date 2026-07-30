// Package target is the single source of truth for cross-compilation targets:
// the GOOS/GOARCH matrix, how each target's binary is named, and - declared
// right in the table - whether each target ships as a distro package.
//
// Both the builder (script/goscript/build) and the packaging scripts consume
// it, so adding an architecture in ONE place flows everywhere, and callers can
// enumerate exactly the targets a given package format ships (on-demand builds).
//
// Package arch names are also kept in the table, so maintainers can see the Go
// build target and every package target name together.
package target

import (
	"fmt"
	"runtime"
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

	// 386
	GO386 string

	// mips
	MIPSSoftFloat bool

	// Whether this row ships as a package for each distro family.
	DEB       bool
	RPM       bool
	ArchLinux bool
	OpenWrt   bool
	Alpine    bool

	DEBArch        string
	DEBArchVariant string
	RPMArch        string
	ArchLinuxArch  string
	OpenWrtArch    []string
	AlpineArch     string
}

// all is the full target matrix. Commented-out entries are kept for the record
// (and pre-tagged with their packaging support for whenever they're enabled).
var all = []Target{
	// darwin (no distro package)
	{GOOS: "darwin", GOARCH: "amd64"},
	{GOOS: "darwin", GOARCH: "arm64"},

	// android (no distro package)
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

	// linux-amd64 (baseline ships for Alpine/OpenWrt; microarch levels are deb/rpm extras)
	{
		GOOS: "linux", GOARCH: "amd64",
		DEB: true, RPM: true, ArchLinux: true, OpenWrt: true, Alpine: true,
		DEBArch: "amd64", DEBArchVariant: "amd64", RPMArch: "x86_64", ArchLinuxArch: "x86_64",
		OpenWrtArch: []string{"x86_64"}, AlpineArch: "x86_64",
	},
	{
		GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 1,
		DEB: true, RPM: true,
		DEBArch: "amd64", DEBArchVariant: "amd64v1", RPMArch: "x86-64-v1",
	},
	{
		GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 2,
		DEB: true, RPM: true,
		DEBArch: "amd64", DEBArchVariant: "amd64v2", RPMArch: "x86-64-v2",
	},
	{
		GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3,
		DEB: true, RPM: true,
		DEBArch: "amd64", DEBArchVariant: "amd64v3", RPMArch: "x86-64-v3",
	},

	// linux-arm
	{
		GOOS: "linux", GOARCH: "arm", GOARMVersion: 5,
		DEB: true, RPM: true, OpenWrt: true,
		DEBArch: "armel", DEBArchVariant: "armel", RPMArch: "armv5tel",
		OpenWrtArch: []string{"arm_arm926ej-s", "arm_cortex-a7", "arm_cortex-a9", "arm_fa526", "arm_xscale"},
	},
	{
		GOOS: "linux", GOARCH: "arm", GOARMVersion: 6,
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "armel", DEBArchVariant: "armel", RPMArch: "armv6hl",
		OpenWrtArch: []string{"arm_arm1176jzf-s_vfp"}, AlpineArch: "armhf",
	},
	{
		GOOS: "linux", GOARCH: "arm", GOARMVersion: 7,
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "armhf", DEBArchVariant: "armhf", RPMArch: "armv7hl",
		OpenWrtArch: []string{
			"arm_cortex-a5_vfpv4", "arm_cortex-a7_neon-vfpv4", "arm_cortex-a7_vfpv4",
			"arm_cortex-a8_vfpv3", "arm_cortex-a9_neon", "arm_cortex-a9_vfpv3-d16",
			"arm_cortex-a15_neon-vfpv4",
		},
		AlpineArch: "armv7",
	},
	{
		GOOS: "linux", GOARCH: "arm64",
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "arm64", DEBArchVariant: "arm64", RPMArch: "aarch64",
		OpenWrtArch: []string{"aarch64_cortex-a53", "aarch64_cortex-a72", "aarch64_cortex-a76", "aarch64_generic"},
		AlpineArch:  "aarch64",
	},

	// linux-mips (only hardfloat ships for deb; rpm/alpine have no mips)
	// mips && mipsle
	{
		GOOS: "linux", GOARCH: "mips", MIPSSoftFloat: true,
		OpenWrt:     true,
		OpenWrtArch: []string{"mips_24kc", "mips_4kec", "mips_mips32"},
	},
	{
		GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: true,
		OpenWrt:     true,
		OpenWrtArch: []string{"mipsel_24kc", "mipsel_74kc", "mipsel_mips32"},
	},
	// mipshf && mipslehf
	{GOOS: "linux", GOARCH: "mips", MIPSSoftFloat: false, DEB: true, DEBArch: "mips", DEBArchVariant: "mips"},
	{
		GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: false,
		DEB: true, OpenWrt: true,
		DEBArch: "mipsel", DEBArchVariant: "mipsel",
		OpenWrtArch: []string{"mipsel_24kc_24kf"},
	},
	{
		GOOS: "linux", GOARCH: "mips64", MIPSSoftFloat: true,
		OpenWrt:     true,
		OpenWrtArch: []string{"mips64_mips64r2", "mips64_octeonplus"},
	},
	{
		GOOS: "linux", GOARCH: "mips64le", MIPSSoftFloat: true,
		OpenWrt:     true,
		OpenWrtArch: []string{"mips64el_mips64r2"},
	},

	// linux-others
	{
		GOOS: "linux", GOARCH: "386",
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "i386", DEBArchVariant: "i386", RPMArch: "i686",
		OpenWrtArch: []string{"i386_pentium4"}, AlpineArch: "x86",
	},
	{
		GOOS: "linux", GOARCH: "386", GO386: "softfloat",
		OpenWrt:     true,
		OpenWrtArch: []string{"i386_pentium-mmx"},
	},
	{
		GOOS: "linux", GOARCH: "loong64",
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "loong64", DEBArchVariant: "loong64", RPMArch: "loongarch64",
		OpenWrtArch: []string{"loongarch64_generic"}, AlpineArch: "loongarch64",
	},
	//{GOOS: "linux", GOARCH: "ppc64", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "ppc64le", Deb: true, RPM: true, Alpine: true, DEBArch: "ppc64el", DEBArchVariant: "ppc64el", RPMArch: "ppc64le", AlpineArch: "ppc64le"},
	{
		GOOS: "linux", GOARCH: "riscv64",
		DEB: true, RPM: true, OpenWrt: true, Alpine: true,
		DEBArch: "riscv64", DEBArchVariant: "riscv64", RPMArch: "riscv64",
		OpenWrtArch: []string{"riscv64_generic"}, AlpineArch: "riscv64",
	},
	//{GOOS: "linux", GOARCH: "s390x", Deb: true, RPM: true, Alpine: true, DEBArch: "s390x", DEBArchVariant: "s390x", RPMArch: "s390x", AlpineArch: "s390x"},
	//{GOOS: "linux", GOARCH: "sparc64", Deb: true},

	// openbsd (uses pkg_add, no distro package here)
	{GOOS: "openbsd", GOARCH: "386"},
	{GOOS: "openbsd", GOARCH: "amd64"},
	{GOOS: "openbsd", GOARCH: "arm", GOARMVersion: 7},
	{GOOS: "openbsd", GOARCH: "arm64"},
	{GOOS: "openbsd", GOARCH: "ppc64"},
	{GOOS: "openbsd", GOARCH: "riscv64"},

	// windows (no distro package)
	{GOOS: "windows", GOARCH: "386"},
	{GOOS: "windows", GOARCH: "amd64"},
	{GOOS: "windows", GOARCH: "arm64"},

	// others (no distro package)
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
	return all
}

// Host returns the baseline target for the current runtime GOOS/GOARCH.
func Host() (Target, bool) {
	for _, t := range all {
		if t.GOOS == runtime.GOOS && t.GOARCH == runtime.GOARCH {
			return t, true
		}
	}
	return Target{}, false
}

func DEBTargets(tgt []Target, goos, goarch string) []Target {
	var filtered []Target
	for _, target := range tgt {
		if target.DEB {
			filtered = append(filtered, target)
		}
	}
	return FilterTargets(filtered, goos, goarch)
}

func RPMTargets(tgt []Target, goos, goarch string) []Target {
	var filtered []Target
	for _, target := range tgt {
		if target.RPM {
			filtered = append(filtered, target)
		}
	}
	return FilterTargets(filtered, goos, goarch)
}

func ArchLinuxTargets(tgt []Target, goos, goarch string) []Target {
	var filtered []Target
	for _, target := range tgt {
		if target.ArchLinux {
			filtered = append(filtered, target)
		}
	}
	return FilterTargets(filtered, goos, goarch)
}

func OpenWrtTargets(tgt []Target, goos, goarch string) []Target {
	var filtered []Target
	for _, target := range tgt {
		if target.OpenWrt {
			filtered = append(filtered, target)
		}
	}
	return FilterTargets(filtered, goos, goarch)
}

func AlpineAPKTargets(tgt []Target, goos, goarch string) []Target {
	var filtered []Target
	for _, target := range tgt {
		if target.Alpine {
			filtered = append(filtered, target)
		}
	}
	return FilterTargets(filtered, goos, goarch)
}

func FilterTargets(tgt []Target, goos, goarch string) []Target {
	if goos == "" && goarch == "" {
		return tgt
	}
	var filtered []Target
	for _, t := range tgt {
		if matchEmptyGlobal(t.GOOS, goos) &&
			matchEmptyGlobal(t.GOARCH, goarch) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (t Target) variantParts() []string {
	parts := []string{t.GOOS, t.GOARCH}

	if t.GOAMD64Version != 0 && t.GOARCH == "amd64" {
		parts = append(parts, fmt.Sprintf("v%d", t.GOAMD64Version))
	}

	switch {
	case t.MIPSSoftFloat && (t.GOARCH == "mips" || t.GOARCH == "mipsle" || t.GOARCH == "mips64" || t.GOARCH == "mips64le"):
		parts = append(parts, "softfloat")
	case t.GOARCH == "mips" || t.GOARCH == "mipsle" || t.GOARCH == "mips64" || t.GOARCH == "mips64le":
		parts = append(parts, "hardfloat")
	}

	if t.GOARMVersion > 0 && t.GOARCH == "arm" {
		parts = append(parts, fmt.Sprintf("v%d", t.GOARMVersion))
	}

	if t.GO386 != "" && t.GOARCH == "386" {
		parts = append(parts, t.GO386)
	}

	return parts
}

func (t Target) BinaryName(base string) string {
	return QualifyName(append([]string{base}, t.variantParts()...)...)
}

func matchEmptyGlobal(raw, c string) bool {
	return raw == c || len(c) == 0 || c == "*"
}

func QualifyName(name ...string) string {
	return strings.Join(name, "-")
}

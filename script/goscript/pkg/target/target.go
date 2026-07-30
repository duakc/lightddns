// Package target is the single source of truth for cross-compilation targets:
// the GOOS/GOARCH matrix, how each target's binary is named, and - declared
// right in the table - whether each target ships as a .deb / .rpm / Arch Linux package.
//
// Both the builder (script/goscript/build) and the packaging scripts consume
// it, so adding an architecture in ONE place flows everywhere, and callers can
// enumerate exactly the targets a given package format ships (on-demand builds).
//
// The Deb/RPM/ArchLinux flags are filled in for every row we know about -
// including the commented-out, not-yet-enabled ones - so whoever enables a rare
// OS/arch later does not have to figure out its packaging support: it is already
// decided here.
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

	// Whether this row ships as a .deb / .rpm / Arch Linux package. Only ONE
	// variant per arch is marked (e.g. baseline amd64, hardfloat mips);
	// microarch/softfloat duplicates stay false. The arch NAME is derived by the
	// package-specific mapping method, not repeated per row.
	DEB       bool
	RPM       bool
	ArchLinux bool
}

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
	{GOOS: "linux", GOARCH: "amd64", DEB: true, RPM: true, ArchLinux: true},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 1, DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 2, DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3, DEB: true, RPM: true},

	// linux-arm (rpm only packages the v7/hardfloat profile)
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 5, DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 6, DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "arm", GOARMVersion: 7, DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "arm64", DEB: true, RPM: true},

	// linux-mips (only hardfloat ships; rpm has no mips)
	// mips && mipsle
	{GOOS: "linux", GOARCH: "mips", MIPSSoftFloat: true},
	{GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: true},
	// mipshf && mipslehf
	{GOOS: "linux", GOARCH: "mips", MIPSSoftFloat: false, DEB: true},
	{GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: false, DEB: true},

	// linux-others
	{GOOS: "linux", GOARCH: "386", DEB: true, RPM: true},
	{GOOS: "linux", GOARCH: "loong64", DEB: true, RPM: true},
	//{GOOS: "linux", GOARCH: "ppc64", Deb: true, RPM: true},
	//{GOOS: "linux", GOARCH: "ppc64le", Deb: true, RPM: true},
	{GOOS: "linux", GOARCH: "riscv64", DEB: true, RPM: true},
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

// OpenWrtTarget is one compile variant and the OpenWrt subtarget arch labels
// that share its binary. OpenWrt ships far more arch labels than Go has arch
// variants: e.g. every arm64 board flavour reuses the single linux/arm64 build.
type OpenWrtTarget struct {
	Target

	// Archs are the OpenWrt package architecture labels (e.g. aarch64_cortex-a53)
	// this binary is published under. One .ipk / .apk is emitted per label.
	Archs []string
}

// openwrtTargets mirrors the sing-box OpenWrt matrix: comprehensive coverage of
// the OpenWrt arch labels, grouped by the Go build variant that produces each.
var openwrtTargets = []OpenWrtTarget{
	{Target{GOOS: "linux", GOARCH: "amd64"}, []string{"x86_64"}},
	{Target{GOOS: "linux", GOARCH: "arm64"}, []string{
		"aarch64_cortex-a53", "aarch64_cortex-a72", "aarch64_cortex-a76", "aarch64_generic",
	}},
	// GO386=sse2 is the toolchain default.
	{Target{GOOS: "linux", GOARCH: "386"}, []string{"i386_pentium4"}},
	{Target{GOOS: "linux", GOARCH: "386", GO386: "softfloat"}, []string{"i386_pentium-mmx"}},
	{Target{GOOS: "linux", GOARCH: "arm", GOARMVersion: 7}, []string{
		"arm_cortex-a5_vfpv4", "arm_cortex-a7_neon-vfpv4", "arm_cortex-a7_vfpv4",
		"arm_cortex-a8_vfpv3", "arm_cortex-a9_neon", "arm_cortex-a9_vfpv3-d16",
		"arm_cortex-a15_neon-vfpv4",
	}},
	{Target{GOOS: "linux", GOARCH: "arm", GOARMVersion: 6}, []string{"arm_arm1176jzf-s_vfp"}},
	{Target{GOOS: "linux", GOARCH: "arm", GOARMVersion: 5}, []string{
		"arm_arm926ej-s", "arm_cortex-a7", "arm_cortex-a9", "arm_fa526", "arm_xscale",
	}},
	{Target{GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: true}, []string{
		"mipsel_24kc", "mipsel_74kc", "mipsel_mips32",
	}},
	{Target{GOOS: "linux", GOARCH: "mipsle"}, []string{"mipsel_24kc_24kf"}}, // hardfloat
	{Target{GOOS: "linux", GOARCH: "mips", MIPSSoftFloat: true}, []string{
		"mips_24kc", "mips_4kec", "mips_mips32",
	}},
	{Target{GOOS: "linux", GOARCH: "mips64", MIPSSoftFloat: true}, []string{
		"mips64_mips64r2", "mips64_octeonplus",
	}},
	{Target{GOOS: "linux", GOARCH: "mips64le", MIPSSoftFloat: true}, []string{"mips64el_mips64r2"}},
	{Target{GOOS: "linux", GOARCH: "riscv64"}, []string{"riscv64_generic"}},
	{Target{GOOS: "linux", GOARCH: "loong64"}, []string{"loongarch64_generic"}},
}

// OpenWrtTargets returns the OpenWrt build variants. Since every variant is a
// linux cross-build, buildAll=false narrows by the host GOARCH only (GOOS is
// irrelevant), so the command is usable from a non-linux dev host.
func OpenWrtTargets(goos, goarch string) []OpenWrtTarget {
	if goos == "" && goarch == "" {
		return append([]OpenWrtTarget(nil), openwrtTargets...)
	}
	var out []OpenWrtTarget
	for _, t := range openwrtTargets {
		if matchEmptyGlobal(t.GOARCH, goarch) &&
			matchEmptyGlobal(t.GOOS, goos) {
			out = append(out, t)
		}
	}
	return out
}

func (t Target) variantParts() []string {
	parts := []string{t.GOOS, t.GOARCH}

	if t.GOAMD64Version != 0 && t.GOARCH == "amd64" {
		parts = append(parts, fmt.Sprintf("v%d", t.GOAMD64Version))
	}

	if t.MIPSSoftFloat && t.GOARCH == "mips" {
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
	return QualifyName(append([]string{base}, t.variantParts()...)...)
}

func (t Target) DEBArchName() string {
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
	default:
		// amd64/arm64/mips/riscv64/loong64/ppc64/s390x/sparc64
		// all map straight through.
		return t.GOARCH
	}
}

func (t Target) DEBArchVariantName() string {
	switch {
	case t.GOARCH == "amd64" && t.GOAMD64Version == 1:
		return "amd64v1"
	case t.GOARCH == "amd64" && t.GOAMD64Version == 2:
		return "amd64v2"
	case t.GOARCH == "amd64" && t.GOAMD64Version == 3:
		return "amd64v3"
	}
	return t.DEBArchName()
}

func (t Target) RPMArchName() string {
	switch t.GOARCH {
	case "amd64":
		switch t.GOAMD64Version {
		case 1:
			return "x86-64-v1"
		case 2:
			return "x86-64-v2"
		case 3:
			return "x86-64-v3"
		}
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "arm":
		switch t.GOARMVersion {
		case 5:
			return "armv5tel"
		case 6:
			return "armv6hl" // or armv6l
		}
		return "armv7hl" // or armv7hnl
	case "386":
		return "i686"
	case "loong64":
		return "loongarch64"
	default:
		// riscv64/ppc64/ppc64le/s390x all map straight through.
		return t.GOARCH
	}
}

func (t Target) ArchLinuxArchName() string {
	if t.GOARCH != "amd64" {
		return t.GOARCH
	}
	return "x86_64"
}

func matchEmptyGlobal(raw, c string) bool {
	return raw == c || len(c) == 0 || c == "*"
}

func QualifyName(name ...string) string {
	return strings.Join(name, "-")
}

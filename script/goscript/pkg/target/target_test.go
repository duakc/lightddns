package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargets(t *testing.T) {
	t.Run("no goos,no goarch", func(t *testing.T) {
		allTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "linux", GOARCH: "arm64"},
			{GOOS: "windows", GOARCH: "amd64"},
		}
		// should contain all
		assert.Equal(t, allTargets, FilterTargets(allTargets, "", ""))
		assert.Equal(t, allTargets, FilterTargets(allTargets, "", "*"))
		assert.Equal(t, allTargets, FilterTargets(allTargets, "*", ""))
		assert.Equal(t, allTargets, FilterTargets(allTargets, "*", "*"))
	})

	t.Run("no goos,goarch", func(t *testing.T) {
		allTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "linux", GOARCH: "arm64"},
			{GOOS: "windows", GOARCH: "amd64"},
		}
		filteredTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "windows", GOARCH: "amd64"},
		}
		assert.Equal(t, filteredTargets, FilterTargets(allTargets, "", "amd64"))
		assert.Equal(t, filteredTargets, FilterTargets(allTargets, "*", "amd64"))
	})

	t.Run("goos,no goarch", func(t *testing.T) {
		allTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "linux", GOARCH: "arm64"},
			{GOOS: "windows", GOARCH: "amd64"},
		}
		filteredTargets := []Target{
			{GOOS: "windows", GOARCH: "amd64"},
		}
		assert.Equal(t, filteredTargets, FilterTargets(allTargets, "windows", ""))
		assert.Equal(t, filteredTargets, FilterTargets(allTargets, "windows", "*"))
	})

	t.Run("goos,goarch", func(t *testing.T) {
		allTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "linux", GOARCH: "arm64"},
			{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3},
			{GOOS: "windows", GOARCH: "amd64"},
		}
		filteredTargets := []Target{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "linux", GOARCH: "amd64", GOAMD64Version: 3},
		}
		assert.Equal(t, filteredTargets, FilterTargets(allTargets, "linux", "amd64"))
	})
}

func TestPackageTargets(t *testing.T) {
	t.Run("openwrt arm64", func(t *testing.T) {
		targets := OpenWrtTargets(All(), "linux", "arm64")
		require.Len(t, targets, 1)
		assert.True(t, targets[0].OpenWrt)
		assert.Contains(t, targets[0].OpenWrtArch, "aarch64_cortex-a53")
		assert.Contains(t, targets[0].OpenWrtArch, "aarch64_generic")
	})

	t.Run("openwrt all", func(t *testing.T) {
		targets := OpenWrtTargets(All(), "linux", "*")
		assert.GreaterOrEqual(t, len(targets), 13)
		for _, tgt := range targets {
			assert.True(t, tgt.OpenWrt)
			assert.NotEmpty(t, tgt.OpenWrtArch)
		}
	})

	t.Run("openwrt mipsle", func(t *testing.T) {
		targets := OpenWrtTargets(All(), "linux", "mipsle")
		require.Len(t, targets, 2)
		assert.Contains(t, targets[0].OpenWrtArch, "mipsel_24kc")
		assert.Contains(t, targets[1].OpenWrtArch, "mipsel_24kc_24kf")
	})

	t.Run("alpine explicit arch", func(t *testing.T) {
		targets := AlpineAPKTargets(All(), "linux", "*")
		for _, tgt := range targets {
			assert.True(t, tgt.Alpine)
			assert.NotEmpty(t, tgt.AlpineArch)
		}
		assertAlpineArch(t, targets, "amd64", 0, "x86_64")
		assertAlpineArch(t, targets, "386", 0, "x86")
		assertAlpineArch(t, targets, "arm64", 0, "aarch64")
		assertAlpineArch(t, targets, "arm", 7, "armv7")
		assertAlpineArch(t, targets, "arm", 6, "armhf")
		assertAlpineArch(t, targets, "riscv64", 0, "riscv64")
		assertAlpineArch(t, targets, "loong64", 0, "loongarch64")
	})
}

func TestPackageArchFields(t *testing.T) {
	t.Run("deb and rpm arch fields", func(t *testing.T) {
		debTargets := DEBTargets(All(), "linux", "amd64")
		require.Len(t, debTargets, 4)
		assert.Equal(t, "amd64", debTargets[0].DEBArch)
		assert.Equal(t, "amd64", debTargets[0].DEBArchVariant)
		assert.Equal(t, "amd64v3", debTargets[3].DEBArchVariant)

		rpmTargets := RPMTargets(All(), "linux", "arm64")
		require.Len(t, rpmTargets, 1)
		assert.Equal(t, "aarch64", rpmTargets[0].RPMArch)
	})

	t.Run("binary variants include softfloat and 386 mode", func(t *testing.T) {
		assert.Equal(t, "lightddns-linux-mipsle-softfloat", Target{GOOS: "linux", GOARCH: "mipsle", MIPSSoftFloat: true}.BinaryName("lightddns"))
		assert.Equal(t, "lightddns-linux-mips64-softfloat", Target{GOOS: "linux", GOARCH: "mips64", MIPSSoftFloat: true}.BinaryName("lightddns"))
		assert.Equal(t, "lightddns-linux-386-softfloat", Target{GOOS: "linux", GOARCH: "386", GO386: "softfloat"}.BinaryName("lightddns"))
	})
}

func assertAlpineArch(t *testing.T, targets []Target, goarch string, goarm int, arch string) {
	t.Helper()
	for _, tgt := range targets {
		if tgt.GOARCH == goarch && tgt.GOARMVersion == goarm {
			assert.Equal(t, arch, tgt.AlpineArch)
			return
		}
	}
	t.Fatalf("missing alpine target %s armv%d", goarch, goarm)
}

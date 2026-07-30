package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

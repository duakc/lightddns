package buildinfo

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemverValue(t *testing.T) {
	t.Run("normalizes valid semantic versions", func(t *testing.T) {
		var value semverValue
		require.NoError(t, value.Set("v1.2.3-rc.1+build.7"))
		assert.Equal(t, semverValue("1.2.3-rc.1+build.7"), value)
	})

	t.Run("rejects invalid versions", func(t *testing.T) {
		var value semverValue
		assert.Error(t, value.Set("not-a-version"))
		assert.Empty(t, value)
	})
}

func TestFlags(t *testing.T) {
	var values flags
	flagSet := flag.NewFlagSet("buildinfo", flag.ContinueOnError)
	values.register(flagSet)

	require.NoError(t, flagSet.Parse([]string{
		"--buildinfo_version", "v2.3.4-beta.1",
		"--buildinfo_branch", "release",
	}))

	assert.Equal(t, Info{Version: "2.3.4-beta.1", Branch: "release"}, values.current())
}

func TestFlagsRejectInvalidVersion(t *testing.T) {
	var values flags
	flagSet := flag.NewFlagSet("buildinfo", flag.ContinueOnError)
	values.register(flagSet)

	assert.Error(t, flagSet.Parse([]string{"--buildinfo_version", "not-a-version"}))
}

func TestFlagsDefaultUnknown(t *testing.T) {
	assert.Equal(t, Info{Version: unknown, Branch: unknown}, (flags{}).current())
}

func TestFlagsUseProvidedMetadata(t *testing.T) {
	assert.Equal(t, Info{Version: "1.2.3", Branch: "release"}, (flags{version: "1.2.3", branch: "release"}).current())
}

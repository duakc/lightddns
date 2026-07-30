package gitver

import (
	"strconv"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestGitVer(T *testing.T) {
	version, err := semver.NewVersion("v0.0.0-alpha.1-1-ga3cc3ad-dirty")
	if err != nil {
		require.NoError(T, err)
	}

	for _, vv := range []string{
		strconv.Itoa(int(version.Major())),
		strconv.Itoa(int(version.Minor())),
		strconv.Itoa(int(version.Patch())),
		version.Prerelease(),
		version.Metadata()} {
		T.Log(vv)
		T.Log("\n")
	}
}

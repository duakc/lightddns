package gitver

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/duakc/lightddns/script/goscript/pkg/common"

	"github.com/Masterminds/semver/v3"
)

// EnvBranch overrides the Git branch stamped into release builds. This is
// needed when CI checks out a tag, which leaves Git in detached HEAD state.
const EnvBranch = "GOSCRIPT_GIT_BRANCH"

var (
	gitVersionOnce, gitBranchOnce, gitLocatableVersionOnce sync.Once

	gitVersion       *semver.Version
	gitVersionString string

	gitBranch           string
	gitLocatableVersion string
)

func Semver(ctx context.Context) *semver.Version {
	Version(ctx)

	copiedSemver := *gitVersion
	return &copiedSemver
}

func Version(ctx context.Context) string {
	gitVersionOnce.Do(func() {
		gitStringVersion := unknown
		if tags := git(ctx, "tag", "--points-at", "HEAD"); tags != "" {
			gitStringVersion = strings.SplitN(tags, "\n", 2)[0] // first tag if several
		} else if hash := git(ctx, "describe", "--tags", "--dirty"); hash != "" {
			gitStringVersion = hash
		}

		parsedSemver, parseErr := semver.NewVersion(gitStringVersion)
		if parseErr != nil {
			common.Fatalf("bad semver: %s: %s", gitStringVersion, parseErr)
		}
		gitVersion = parsedSemver
		gitVersionString = parsedSemver.String()
	})

	return gitVersionString
}

func Branch(ctx context.Context) string {
	gitBranchOnce.Do(func() {
		gitBranch = branch(ctx)
	})
	return gitBranch
}

func branch(ctx context.Context) string {
	if b := os.Getenv(EnvBranch); b != "" {
		return b
	}
	if b := git(ctx, "branch", "--show-current"); b != "" {
		return b
	}
	return unknown
}

func LocatableVersion(ctx context.Context) string {
	gitLocatableVersionOnce.Do(func() {
		gitLocatableVersion = unknown
		if tags := git(ctx, "tag", "--points-at", "HEAD"); tags != "" {
			gitLocatableVersion = strings.SplitN(tags, "\n", 2)[0]
		} else if hash := git(ctx, "rev-parse", "HEAD"); hash != "" {
			gitLocatableVersion = hash
		}
	})
	return gitLocatableVersion
}

const unknown = "(unknown)"

func git(ctx context.Context, args ...string) string {
	out, _ := common.CommandCapture(ctx, common.Cmd{Name: "git", Args: args})
	return out
}

package gitver

import (
	"context"
	"strings"
	"sync"

	"github.com/duakc/lightddns/script/goscript/pkg/common"

	"github.com/Masterminds/semver/v3"
)

var (
	gitVersionOnce, gitBranchOnce sync.Once

	gitVersion       *semver.Version
	gitVersionString string

	gitBranch string
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
		gitBranch = unknown
		if b := git(ctx, "branch", "--show-current"); b != "" {
			gitBranch = b
		}
	})
	return gitBranch
}

const unknown = "(unknown)"

func git(ctx context.Context, args ...string) string {
	out, _ := common.CommandCapture(ctx, common.Cmd{Name: "git", Args: args})
	return out
}

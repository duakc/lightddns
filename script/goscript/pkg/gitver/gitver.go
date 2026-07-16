package gitver

import (
	"context"
	"strings"
	"sync"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
)

var (
	gitVersionOnce, gitBranchOnce sync.Once

	gitVersion, gitBranch string
)

func Version(ctx context.Context) string {
	gitVersionOnce.Do(func() {
		gitVersion = unknown
		if tags := git(ctx, "tag", "--points-at", "HEAD"); tags != "" {
			gitVersion = strings.SplitN(tags, "\n", 2)[0] // first tag if several
		} else if hash := git(ctx, "rev-parse", "--short", "HEAD"); hash != "" {
			gitVersion = hash
		}
	})
	return gitVersion
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

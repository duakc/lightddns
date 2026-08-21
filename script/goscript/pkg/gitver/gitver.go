package gitver

import (
	"context"
	"strings"
	"sync"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
)

var (
	gitLocatableVersionOnce sync.Once
	gitLocatableVersion     string
)

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

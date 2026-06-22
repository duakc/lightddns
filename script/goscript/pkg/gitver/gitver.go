package gitver

import (
	"context"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
)

func Version(ctx context.Context) string {
	if tags := git(ctx, "tag", "--points-at", "HEAD"); tags != "" {
		return strings.SplitN(tags, "\n", 2)[0] // first tag if several
	}
	if hash := git(ctx, "rev-parse", "--short", "HEAD"); hash != "" {
		return hash
	}
	return unknown
}

func Branch(ctx context.Context) string {
	if b := git(ctx, "branch", "--show-current"); b != "" {
		return b
	}
	return unknown
}

const unknown = "(unknown)"

func git(ctx context.Context, args ...string) string {
	out, _ := common.CommandCapture(ctx, common.Cmd{Name: "git", Args: args})
	return out
}

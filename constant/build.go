package constant

import (
	"strings"
	"sync"
)

const unknown = "(unknown)"

var (
	Version = unknown
	Branch  = unknown

	Tags = unknown
)

var (
	tagSplitOnce sync.Once
	tagsArray    []string
)

func TagList() []string {
	tagSplitOnce.Do(func() {
		tagsArray = strings.Split(Tags, ",")
	})

	return tagsArray
}

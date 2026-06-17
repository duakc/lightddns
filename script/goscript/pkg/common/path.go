package common

import (
	"os"
	"path/filepath"
)

const (
	// Note: sync with Makefile.
	draftRoot   = "draft"
	buildRoot   = "build"
	releaseRoot = "release"
)

var buildDraft = filepath.Join(buildRoot, draftRoot)

// BuildDir returns a path under the build output dir, creating it so callers
// don't each have to MkdirAll.
func BuildDir(parts ...string) string {
	return ensure(filepath.Join(append([]string{buildRoot}, parts...)...))
}

// BuildDraftDir returns a path under build/draft (local scratch), creating it.
func BuildDraftDir(parts ...string) string {
	return ensure(filepath.Join(append([]string{buildDraft}, parts...)...))
}

// ReleaseDir returns a path under release/ (existing payload; not created).
func ReleaseDir(parts ...string) string {
	return filepath.Join(append([]string{releaseRoot}, parts...)...)
}

func ensure(dir string) string {
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

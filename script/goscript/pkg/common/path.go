package common

import (
	"os"
	"path/filepath"
)

var (
	// Note: sync with Makefile.
	draftRoot   = getEnvOrDefault("GOSCRIPT_DRAFT_SUB_DIR", "draft")
	buildRoot   = getEnvOrDefault("GOSCRIPT_BUILD_DIR", "build")
	releaseRoot = getEnvOrDefault("GOSCRIPT_RELEASE_DIR", "release")
)

var buildDraft = filepath.Join(buildRoot, draftRoot)

func BuildDir(parts ...string) string {
	return ensure(filepath.Join(append([]string{buildRoot}, parts...)...))
}

func BuildDraftDir(parts ...string) string {
	return ensure(filepath.Join(append([]string{buildDraft}, parts...)...))
}

func ReleaseDir(parts ...string) string {
	return filepath.Join(append([]string{releaseRoot}, parts...)...)
}

func ensure(dir string) string {
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

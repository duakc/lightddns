// Package buildprofile is the build-system's config layer: it loads the release
// build profile(s) from a YAML file so the "common" build parameters (tags, the
// arch matrix, output naming) live in data rather than in the Makefile.
// Per-invocation knobs (version, branch, ...) stay as flags/env vars.
//
// The file is resolved from $RELEASE_BUILD_PROFILE (a path, relative to the
// working dir) if set, otherwise <release dir>/build.yaml. It is a YAML stream:
// each document is one independent build profile.
package buildprofile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/duakc/lightddns/script/goscript/pkg/common"

	goyaml "github.com/goccy/go-yaml"
)

// EnvProfilePath overrides the profile location with a path (relative to the
// working dir, or absolute).
const EnvProfilePath = "RELEASE_BUILD_PROFILE"

// Profile is one default build-parameter set (one document in the YAML stream).
// It holds the params common to a build; the target architecture is chosen per
// invocation (a build flag), not stored here. Add future common knobs here.
type Profile struct {
	Enabled bool `yaml:"enabled"`

	// DefaultTags are the Go build tags applied to every artifact (e.g. which
	// providers to compile in).
	DefaultTags []string `yaml:"defaultTags"`

	// BuildName names the output tree: empty -> build/bin, else
	// build/build_<buildName>/bin. Must be unique across enabled profiles.
	BuildName string `yaml:"buildName"`
}

// Path is the resolved profile location: $RELEASE_BUILD_PROFILE if set, else
// <release dir>/build.yaml.
func Path() string {
	if p := os.Getenv(EnvProfilePath); p != "" {
		return p
	}
	return common.ReleaseDir("build.yaml")
}

// Load reads every profile document from Path(). A missing file is not an error
// (returns nil), so callers can fall back to their flag defaults.
func Load() ([]Profile, error) {
	return LoadFile(Path())
}

// LoadFile reads every profile document from path.
func LoadFile(path string) ([]Profile, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var profiles []Profile
	dec := goyaml.NewDecoder(f)
	for {
		var p Profile
		if err := dec.Decode(&p); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// Enabled returns only the profiles with enabled: true.
func Enabled(profiles []Profile) []Profile {
	var out []Profile
	for _, p := range profiles {
		if p.Enabled {
			out = append(out, p)
		}
	}
	return out
}

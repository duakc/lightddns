// Package buildinfo supplies build metadata shared by binaries and packages.
package buildinfo

import (
	"flag"
	"fmt"

	"github.com/duakc/lightddns/script/goscript/pkg/common"

	"github.com/Masterminds/semver/v3"
)

var defaultFlags flags

const unknown = "(unknown)"

type Info struct {
	Version string
	Branch  string
}

func RegisterFlags(flagSet *flag.FlagSet) {
	defaultFlags.register(flagSet)
}

func Current() Info {
	return defaultFlags.current()
}

func Version() string {
	return Current().Version
}

func Branch() string {
	return Current().Branch
}

func Semver() *semver.Version {
	version := Version()
	parsed, err := semver.NewVersion(version)
	if err != nil {
		common.Fatalf("bad semver: %s: %s", version, err)
	}
	return parsed
}

type flags struct {
	version semverValue
	branch  string
}

func (f *flags) register(flagSet *flag.FlagSet) {
	flagSet.Var(&f.version, "buildinfo_version", "version to stamp into build artifacts; must be a semantic version")
	flagSet.StringVar(&f.branch, "buildinfo_branch", unknown, "source branch to stamp into build metadata")
}

func (f flags) current() Info {
	info := Info{Version: unknown, Branch: unknown}
	if f.version != "" {
		info.Version = string(f.version)
	}
	if f.branch != "" {
		info.Branch = f.branch
	}
	return info
}

type semverValue string

func (v *semverValue) Set(value string) error {
	parsed, err := semver.NewVersion(value)
	if err != nil {
		return fmt.Errorf("must be a semantic version: %w", err)
	}
	*v = semverValue(parsed.String())
	return nil
}

func (v semverValue) String() string {
	if v == "" {
		return unknown
	}
	return string(v)
}

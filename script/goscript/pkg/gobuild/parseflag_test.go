package gobuild

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQuotedValues(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{name: "empty", value: "", expected: nil},
		{name: "one value", value: "A=1", expected: []string{"A=1"}},
		{
			name:  "quoted values",
			value: `-X "example.com/project.Version=1.2.3" -X "example.com/project.Tags=debug"`,
			expected: []string{
				"-X", "example.com/project.Version=1.2.3",
				"-X", "example.com/project.Tags=debug",
			},
		},
		{
			name:     "value containing a space",
			value:    `A=1 "B=two words"`,
			expected: []string{"A=1", "B=two words"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseQuotedValues(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}

	_, err := parseQuotedValues(`"unterminated`)
	assert.Error(t, err)
}

func TestBuildFlags(t *testing.T) {
	var values buildFlags
	flagSet := flag.NewFlagSet("gobuild", flag.ContinueOnError)
	values.register(flagSet)

	require.NoError(t, flagSet.Parse([]string{
		"--gobuild_workdir", "./cmd/lightddns",
		"--gobuild_output", "build/bin",
		"--gobuild_binary_name", "lightddns",
		"--gobuild_tags", "freebuf_low_mem, debug ,",
		"--gobuild_env", `CGO_ENABLED=1 "SOURCE=release build"`,
		"--gobuild_ldflags", `-X "example.com/project.Version=1.2.3"`,
		"--gobuild_extra_args", `-mod=readonly "-gcflags=all=-N -l"`,
		"--gobuild_qualified",
	}))

	assert.Equal(t, Params{
		WorkingDir: "./cmd/lightddns",
		OutputDir:  "build/bin",
		BinaryName: "lightddns",
		ExtraTags:  []string{"freebuf_low_mem", "debug"},
		ExtraEnv:   []string{"CGO_ENABLED=1", "SOURCE=release build"},
		LDFlags:    []string{"-X", "example.com/project.Version=1.2.3"},
		ExtraArgs:  []string{"-mod=readonly", "-gcflags=all=-N -l"},
		Qualified:  true,
	}, values.toParams())
}

func TestJoinLDFlagsPreservesWhitespace(t *testing.T) {
	joined, err := joinLDFlags([]string{
		"-X",
		"github.com/duakc/lightddns/infra/netx/httpx.DefaultUserAgent=lightddns/test (go1.27.0)",
		"-s",
	})
	require.NoError(t, err)
	assert.Equal(t, "-X 'github.com/duakc/lightddns/infra/netx/httpx.DefaultUserAgent=lightddns/test (go1.27.0)' -s", joined)
}

func TestJoinLDFlagsRejectsAmbiguousQuotes(t *testing.T) {
	_, err := joinLDFlags([]string{"value with 'single' and \"double\" quotes"})
	require.Error(t, err)
}

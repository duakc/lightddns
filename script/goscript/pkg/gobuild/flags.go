package gobuild

import (
	"encoding/csv"
	"flag"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
)

var defaultBuildFlags buildFlags

func RegisterBuildFlags(flagSet *flag.FlagSet) {
	defaultBuildFlags.register(flagSet)
}

func DefaultParams() Params {
	return defaultBuildFlags.toParams()
}

type buildFlags struct {
	params Params

	extraTags    string
	extraEnv     string
	extraLDFlags string
	extraArgs    string
}

func (f *buildFlags) register(flagSet *flag.FlagSet) {
	flagSet.StringVar(&f.params.WorkingDir,
		buildParamFlagName("workdir"), "", "package directory passed to go build")
	flagSet.StringVar(&f.params.OutputDir,
		buildParamFlagName("output"), "", "directory for built binaries")
	flagSet.StringVar(&f.params.BinaryName,
		buildParamFlagName("binary_name"), "", "base name for built binaries")
	flagSet.StringVar(&f.extraTags,
		buildParamFlagName("tags"), "", "comma-separated build tags")
	flagSet.StringVar(&f.extraEnv,
		buildParamFlagName("env"), "", "space-separated environment assignments; use double quotes for values containing spaces")
	flagSet.StringVar(&f.extraLDFlags,
		buildParamFlagName("ldflags"), "", "space-separated linker flags; use double quotes for values containing spaces")
	flagSet.BoolVar(&f.params.Qualified,
		buildParamFlagName("qualified"), false, "write platform-qualified binary names")
	flagSet.StringVar(&f.extraArgs,
		buildParamFlagName("extra_args"), "", "additional go build arguments; use double quotes for values containing spaces")
}

func (f buildFlags) toParams() Params {
	p := f.params
	p.ExtraTags = splitCommaValues(f.extraTags)
	p.LDFlags = parseQuotedValue("ldflags", f.extraLDFlags)
	p.ExtraEnv = parseQuotedValue("env", f.extraEnv)
	p.ExtraArgs = parseQuotedValue("extra_args", f.extraArgs)
	return p
}

func buildParamFlagName(n string) string {
	return "gobuild_" + n
}

func splitCommaValues(value string) []string {
	var values []string
	for _, value := range strings.Split(value, ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func parseQuotedValue(target, value string) []string {
	values, err := parseQuotedValues(value)
	if err != nil {
		common.Fatalf("bad %s: %s: %v", target, value, err)
	}
	return values
}

func parseQuotedValues(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parser := csv.NewReader(strings.NewReader(value))
	parser.Comma = ' '
	parser.FieldsPerRecord = -1
	parser.TrimLeadingSpace = true
	return parser.Read()
}

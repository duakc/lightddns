package version

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"

	"github.com/duakc/mt"
	"github.com/duakc/mt/debug"

	goyaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var outputMethod string

var Command = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run:   entry,
}

func init() {
	Command.Flags().StringVarP(&outputMethod, "format", "f", "plain", "Output format: json, yaml, or plain (default)")
}

type Info struct {
	Name       string
	Version    string
	Branch     string
	Tags       []string
	Debug      bool
	Datasource []string
	Provider   []string
	Services   []string
}

func entry(cmd *cobra.Command, args []string) {
	I := NewInfo()
	switch strings.ToLower(outputMethod) {
	case "json":
		I.JSON()
	case "yaml", "yml":
		I.YAML()
	case "plain":
		I.Plain()
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unknown output method: %s, use plain as default\n", outputMethod)
		I.Plain()
	}
}

func NewInfo() Info {
	datasourceTypes := adapter.DatasourceRegister.Types()
	providerTypes := adapter.ProviderRegister.Types()
	serviceTypes := adapter.ServiceRegistry.Types()

	sort.Strings(datasourceTypes)
	sort.Strings(providerTypes)
	sort.Strings(serviceTypes)

	return Info{
		Name:    constpkg.Project,
		Version: constpkg.Version,
		Branch:  constpkg.Branch,
		Tags:    displayTags(constpkg.TagList()),

		Debug:      debug.Enabled,
		Datasource: datasourceTypes,
		Provider:   providerTypes,
		Services:   serviceTypes,
	}
}

// displayTags drops provider_with_* tags: which providers are compiled in is
// already reported by the Provider list.
func displayTags(tags []string) []string {
	var out []string
	for _, t := range tags {
		if strings.HasPrefix(t, "provider_with_") {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (I Info) YAML() {
	data := mt.Must(goyaml.Marshal(&I))
	_, _ = os.Stderr.Write(data)
}

func (I Info) JSON() {
	data := mt.Must(json.Marshal(&I))
	_, _ = os.Stderr.Write(data)
}

func (I Info) Plain() {
	var b strings.Builder

	fmt.Fprintf(&b, "%s %s", I.Name, I.Version)
	if I.Branch != "" {
		fmt.Fprintf(&b, " (branch %s)", I.Branch)
	}
	fmt.Fprintf(&b, " debug=%t\n", I.Debug)

	line := func(key, value string) {
		if value == "" {
			return
		}
		fmt.Fprintf(&b, "%s: %s\n", key, value)
	}
	line("tags", strings.Join(I.Tags, ","))
	line("datasources", strings.Join(I.Datasource, ","))
	line("providers", strings.Join(I.Provider, ","))
	line("services", strings.Join(I.Services, ","))

	_, _ = os.Stderr.WriteString(b.String())
}

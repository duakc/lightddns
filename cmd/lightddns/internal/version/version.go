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

		Debug:      debug.Enabled,
		Datasource: datasourceTypes,
		Provider:   providerTypes,
		Services:   serviceTypes,
	}
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
	const temp = `%s: Version: %s, Branch: %s, Debug: %t
Supported Datasource: %s
Supported Provider: %s
Supported Services: %s
`
	_, _ = fmt.Fprintf(os.Stderr, temp, I.Name,
		I.Version, I.Branch, I.Debug,
		strings.Join(I.Datasource, ","),
		strings.Join(I.Provider, ","),
		strings.Join(I.Services, ","))
}

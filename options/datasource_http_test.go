package options_test

import (
	"testing"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/options"

	"github.com/stretchr/testify/require"
)

func TestHTTPDatasourceCommonServiceExampleOptions(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantURL   string
		wantJQ    bool
		wantRegex bool
	}{
		{
			name: "ipinfo.io JSON",
			yaml: `
type: http
name: ipinfo
url: https://ipinfo.io
match:
  jq: ".ip"
`,
			wantURL: "https://ipinfo.io",
			wantJQ:  true,
		},
		{
			name: "api.ip.sb plain text",
			yaml: `
type: http
name: ip-sb
url: https://api.ip.sb/ip
`,
			wantURL: "https://api.ip.sb/ip",
		},
		{
			name: "api.ip.sb/jsonip JSON",
			yaml: `
type: http
name: ipify-json
url: https://api.ip.sb/jsonip
match:
  jq: ".ip"
`,
			wantURL: "https://api.ip.sb/jsonip",
			wantJQ:  true,
		},
		{
			name: "myip.ipip.net regex",
			yaml: `
type: http
name: ipip
url: https://myip.ipip.net
match:
  regex: "当前 IP：\\s*(.+?)\\s*来自于："
`,
			wantURL:   "https://myip.ipip.net",
			wantRegex: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opt options.HTTPDatasourceOption
			require.NoError(t, badyaml.Unmarshal([]byte(tt.yaml), &opt))

			require.Equal(t, "http", opt.Type)
			require.Equal(t, tt.wantURL, opt.URL.Raw)
			require.NotNil(t, opt.URL.URL)
			require.Equal(t, tt.wantJQ, opt.Match.JQ != nil)
			require.Equal(t, tt.wantRegex, opt.Match.Regexp != nil)
		})
	}
}

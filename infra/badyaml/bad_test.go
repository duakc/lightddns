package badyaml

import (
	"testing"
	"time"

	goyaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestListable(t *testing.T) {
	type schema struct {
		Payload Listable[string] `yaml:"payload"`
	}

	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "block list", input: "payload:\n  - e1\n  - \"e2\"\n", expected: []string{"e1", "e2"}},
		{name: "flow list", input: "payload: ['e1','e2','e3']", expected: []string{"e1", "e2", "e3"}},
		{name: "single scalar", input: "payload: e1", expected: []string{"e1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var vv schema
			require.NoError(t, goyaml.Unmarshal([]byte(c.input), &vv))
			assert.Equal(t, c.expected, vv.Payload.Value)
		})
	}
}

// TestHTTPMethod_UnmarshalYAML covers each YAML scalar style. The previous
// implementation used byte-level UnquoteString, which would mishandle YAML
// escape sequences (e.g. embedded "\n") and block scalars — these cases now
// exercise that path.
func TestHTTPMethod_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Method HTTPMethod `yaml:"method"`
	}

	cases := []struct {
		name     string
		input    string
		expected HTTPMethod
		wantErr  bool
	}{
		{name: "plain", input: "method: GET", expected: "GET"},
		{name: "double quoted", input: `method: "GET"`, expected: "GET"},
		{name: "single quoted", input: "method: 'POST'", expected: "POST"},
		{name: "lowercase upcased", input: "method: get", expected: "GET"},
		{name: "RFC2324 BREW", input: "method: brew", expected: "BREW"},
		{name: "empty", input: `method: ""`, expected: ""},
		{name: "folded block", input: "method: >-\n  GET\n", expected: "GET"},
		{name: "literal block", input: "method: |-\n  POST\n", expected: "POST"},
		{name: "invalid method", input: "method: WALK", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			err := goyaml.Unmarshal([]byte(c.input), &s)
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.expected, s.Method)
		})
	}
}

func TestDuration_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Interval Duration `yaml:"interval"`
	}

	cases := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{name: "plain", input: "interval: 5s", expected: 5 * time.Second},
		{name: "double quoted", input: `interval: "5s"`, expected: 5 * time.Second},
		{name: "single quoted", input: "interval: '500ms'", expected: 500 * time.Millisecond},
		{name: "compound", input: `interval: "1h30m"`, expected: 90 * time.Minute},
		{name: "empty defaults to zero", input: `interval: ""`, expected: 0},
		{name: "invalid duration", input: "interval: foo", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			err := goyaml.Unmarshal([]byte(c.input), &s)
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, Duration(c.expected), s.Interval)
		})
	}
}

func TestLogLevel_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Level LogLevel `yaml:"level"`
	}

	cases := []struct {
		name     string
		input    string
		expected zapcore.Level
		wantErr  bool
	}{
		{name: "plain", input: "level: debug", expected: zapcore.DebugLevel},
		{name: "double quoted", input: `level: "info"`, expected: zapcore.InfoLevel},
		{name: "single quoted", input: "level: 'warn'", expected: zapcore.WarnLevel},
		{name: "uppercase", input: "level: ERROR", expected: zapcore.ErrorLevel},
		{name: "invalid level", input: "level: chatty", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			err := goyaml.Unmarshal([]byte(c.input), &s)
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, LogLevel(c.expected), s.Level)
		})
	}
}

func TestURL_UnmarshalYAML(t *testing.T) {
	type schema struct {
		URL URL `yaml:"url"`
	}

	cases := []struct {
		name        string
		input       string
		expectedRaw string
		expectedSch string
	}{
		{
			name:        "plain",
			input:       "url: https://api.ip.sb/ip",
			expectedRaw: "https://api.ip.sb/ip",
			expectedSch: "https",
		},
		{
			name:        "double quoted preserves query",
			input:       `url: "https://api.ip.sb/jsonip"`,
			expectedRaw: "https://api.ip.sb/jsonip",
			expectedSch: "https",
		},
		{
			name:        "single quoted with special chars",
			input:       `url: 'http://example.com/a?b=c&d=e'`,
			expectedRaw: "http://example.com/a?b=c&d=e",
			expectedSch: "http",
		},
		{
			name:        "ipv6 literal host",
			input:       `url: "http://[::1]:8080/"`,
			expectedRaw: "http://[::1]:8080/",
			expectedSch: "http",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			require.NoError(t, goyaml.Unmarshal([]byte(c.input), &s))
			assert.Equal(t, c.expectedRaw, s.URL.Raw)
			require.NotNil(t, s.URL.URL)
			assert.Equal(t, c.expectedSch, s.URL.URL.Scheme)
		})
	}
}

func TestDomainName_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Name DomainName `yaml:"name"`
	}

	cases := []struct {
		name     string
		input    string
		expected DomainName
		wantErr  bool
	}{
		{name: "plain", input: "name: example.com", expected: "example.com"},
		{name: "double quoted", input: `name: "sub.example.com"`, expected: "sub.example.com"},
		{name: "single quoted", input: "name: 'xn--bcher-kva.example'", expected: "xn--bcher-kva.example"},
		{name: "invalid", input: "name: not a domain!", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			err := goyaml.Unmarshal([]byte(c.input), &s)
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.expected, s.Name)
		})
	}
}

// TestHTTPHeader_UnmarshalYAML verifies that values returned by the YAML
// parser are kept verbatim. The previous implementation re-applied
// UnquoteString to strings the parser had already produced, which would strip
// quote characters that were part of legitimate header values.
func TestHTTPHeader_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Headers HTTPHeader `yaml:"headers"`
	}

	cases := []struct {
		name  string
		input string
		check func(t *testing.T, h HTTPHeader)
	}{
		{
			name:  "single value",
			input: "headers:\n  User-Agent: Lightddns/stable\n",
			check: func(t *testing.T, h HTTPHeader) {
				assert.Equal(t, "Lightddns/stable", h.Header.Get("User-Agent"))
			},
		},
		{
			name:  "list value",
			input: "headers:\n  X-Custom:\n    - a\n    - b\n",
			check: func(t *testing.T, h HTTPHeader) {
				assert.Equal(t, []string{"a", "b"}, h.Header.Values("X-Custom"))
			},
		},
		{
			name:  "value with literal double quotes preserved",
			input: `headers:` + "\n  X-Token: \"\\\"abc\\\"\"\n",
			check: func(t *testing.T, h HTTPHeader) {
				assert.Equal(t, `"abc"`, h.Header.Get("X-Token"))
			},
		},
		{
			name:  "bearer token quoted",
			input: `headers:` + "\n  Authorization: \"Bearer xyz\"\n",
			check: func(t *testing.T, h HTTPHeader) {
				assert.Equal(t, "Bearer xyz", h.Header.Get("Authorization"))
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			require.NoError(t, goyaml.Unmarshal([]byte(c.input), &s))
			c.check(t, s.Headers)
		})
	}
}

func TestStringOrNumber_UnmarshalYAML(t *testing.T) {
	type schema struct {
		Value StringOrNumber `yaml:"value"`
	}

	cases := []struct {
		name     string
		input    string
		expected StringOrNumber
	}{
		{name: "bare integer", input: "value: 42", expected: StringOrNumber{Num: 42}},
		{name: "negative integer", input: "value: -7", expected: StringOrNumber{Num: -7}},
		// goccy/go-yaml is lenient: a quoted scalar that parses as int still coerces into int64.
		// This matches the prior UnquoteString+ParseInt behavior.
		{name: "quoted digits coerce to int", input: `value: "42"`, expected: StringOrNumber{Num: 42}},
		{name: "plain string", input: "value: eth0", expected: StringOrNumber{Str: "eth0"}},
		{name: "quoted string", input: `value: "eth0"`, expected: StringOrNumber{Str: "eth0"}},
		{name: "numeric-looking name stays string", input: "value: eth42", expected: StringOrNumber{Str: "eth42"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			require.NoError(t, goyaml.Unmarshal([]byte(c.input), &s))
			assert.Equal(t, c.expected, s.Value)
		})
	}
}

func TestDualStack(t *testing.T) {
	type schema struct {
		URL  DualStack[URL]    `yaml:"url"`
		JSON DualStack[string] `yaml:"json"`
	}

	cases := []struct {
		name  string
		input string
		check func(t *testing.T, s schema)
	}{
		{
			name: "single value applies to both stacks",
			input: `
url: https://api.ip.sb/ip
json: ".ip"
`,
			check: func(t *testing.T, s schema) {
				assert.Equal(t, "https://api.ip.sb/ip", s.URL.IPv4.Raw)
				assert.Equal(t, "https://api.ip.sb/ip", s.URL.IPv6.Raw)
				assert.Equal(t, ".ip", s.JSON.IPv4)
				assert.Equal(t, ".ip", s.JSON.IPv6)
			},
		},
		{
			name: "object form sets each stack separately",
			input: `
url:
  ipv4: https://api.ipify.org
  ipv6: https://api6.ipify.org
json:
  ipv4: ".ipv4"
  ipv6: ".ipv6"
`,
			check: func(t *testing.T, s schema) {
				assert.Equal(t, "https://api.ipify.org", s.URL.IPv4.Raw)
				assert.Equal(t, "https://api6.ipify.org", s.URL.IPv6.Raw)
				assert.Equal(t, ".ipv4", s.JSON.IPv4)
				assert.Equal(t, ".ipv6", s.JSON.IPv6)
			},
		},
		{
			name: "object form allows missing stack",
			input: `
url:
  ipv4: https://api.ipify.org
`,
			check: func(t *testing.T, s schema) {
				assert.Equal(t, "https://api.ipify.org", s.URL.IPv4.Raw)
				assert.Equal(t, "", s.URL.IPv6.Raw)
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s schema
			require.NoError(t, goyaml.Unmarshal([]byte(c.input), &s))
			c.check(t, s)
		})
	}
}

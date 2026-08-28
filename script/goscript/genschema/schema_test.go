package genschema

import (
	"encoding/json"
	"fmt"
	"testing"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/script/goscript/pkg/jsonschema"

	"github.com/stretchr/testify/require"
)

func TestGeneratedSchemaUsesRuntimeEnums(t *testing.T) {
	datasources := datasourceSchema()
	command := findVariant(t, datasources, constpkg.DatasourceTypeCommand)
	require.Equal(t, 1, *command.Properties["cmd"].AnyOf[0].MinLength)
	require.Equal(t, 1, *command.Properties["cmd"].AnyOf[1].MinItems)
	require.Equal(t,
		[]string{"none", "stdout", "stderr", "all"},
		enumStrings(command.Properties["output"]))
	require.JSONEq(t, `"none"`, string(command.Properties["output"].Default))
	require.Equal(t,
		[]string{"stdout", "stderr", "all"},
		enumStrings(command.Properties["capture"]))
	require.JSONEq(t, `"stdout"`, string(command.Properties["capture"].Default))

	logSchema := mustFor[options.LogOption]()
	require.Equal(t,
		[]string{"debug", "info", "warn", "warning", "error", "dpanic", "panic", "fatal"},
		enumStrings(logSchema.Properties["level"]))
	require.JSONEq(t, `"info"`, string(logSchema.Properties["level"].Default))
}

func TestGeneratedSchemaUsesConfigurationShapes(t *testing.T) {
	datasources := datasourceSchema()

	http := findVariant(t, datasources, constpkg.DatasourceTypeHTTP)
	require.Equal(t, `^https?://.+$`, http.Properties["url"].Pattern)
	require.JSONEq(t, `"GET"`, string(http.Properties["method"].Default))
	require.JSONEq(t, `"prefer_ipv6"`,
		string(http.Properties["connect"].Properties["dialStrategy"].Default))
	headers := http.Properties["headers"]
	require.Equal(t, JSONTypeObject, headers.Type)
	require.NotNil(t, headers.AdditionalProperties)
	require.Len(t, headers.AdditionalProperties.AnyOf, 2)

	filter := findVariant(t, datasources, constpkg.DatasourceGroupFilter)
	require.Equal(t, JSONTypeArray, filter.Properties["datasources"].Type)
	prefixes := filter.Properties["rules"].Items.Properties["prefixes"]
	require.Equal(t, JSONTypeString, prefixes.Items.Type)

	netlink := findVariant(t, datasources, constpkg.DatasourceTypeNetlink)
	require.Equal(t, 1, *netlink.Properties["ifName"].MinLength)
	require.Equal(t, float64(1), *netlink.Properties["ifIndex"].Minimum)
	require.Len(t, netlink.AnyOf, 2)

	require.Equal(t, 1, *filter.Properties["rules"].MinItems)
	require.Equal(t, JSONTypeArray, filter.Properties["rules"].Type)

	dns := http.Properties["dns"]
	require.Len(t, dns.AnyOf, 3)
	dnsObject := dns.AnyOf[2]
	require.Equal(t, []string{"system", "tls"}, enumStrings(dnsObject.Properties["type"]))
	require.JSONEq(t, `"system"`, string(dnsObject.Properties["type"].Default))
	require.Len(t, dnsObject.AllOf, 1)
	require.Equal(t, []string{"server"}, dnsObject.AllOf[0].Then.Required)
}

func TestGeneratedSchemaIncludesServiceDefaults(t *testing.T) {
	services := servicesSchema()

	ipserver := findVariant(t, services, constpkg.ServiceTypeIPServer)
	require.JSONEq(t, `9002`, string(ipserver.Properties["port"].Default))
	require.JSONEq(t, `"/"`, string(ipserver.Properties["path"].Default))

	prometheus := findVariant(t, services, constpkg.ServiceTypePrometheus)
	require.JSONEq(t, `9001`, string(prometheus.Properties["port"].Default))
	require.JSONEq(t, `"/metrics"`, string(prometheus.Properties["path"].Default))
}

func TestGeneratedSchemaProvidesProviderLineCompletions(t *testing.T) {
	providers := providerSchema()

	aliyun := findVariant(t, providers, constpkg.ProviderTypeAliyun)
	aliyunLines := aliyun.Properties["lines"]
	require.Len(t, aliyunLines.AnyOf, 2)
	aliyunLineValues := aliyunLines.AnyOf[0]
	require.Len(t, aliyunLineValues.AnyOf, 2)
	require.Equal(t,
		[]string{"default", "telecom", "unicom", "mobile", "oversea", "edu", "drpeng"},
		enumStrings(aliyunLineValues.AnyOf[0]))
	require.Equal(t, JSONTypeString, aliyunLineValues.AnyOf[1].Type)
	require.Equal(t, 1, *aliyunLineValues.AnyOf[1].MinLength)
	require.Equal(t, JSONTypeArray, aliyunLines.AnyOf[1].Type)
	require.Equal(t, 1, *aliyunLines.AnyOf[1].Items.MinLength)
	require.Equal(t, aliyunLineValues.AnyOf[0].Enum, aliyunLines.AnyOf[1].Items.AnyOf[0].Enum)

	tencent := findVariant(t, providers, constpkg.ProviderTypeTencentCloud)
	tencentLines := tencent.Properties["lines"]
	require.Len(t, tencentLines.AnyOf, 2)
	tencentLineValues := tencentLines.AnyOf[0]
	require.Len(t, tencentLineValues.AnyOf, 2)
	require.Equal(t,
		[]string{"默认", "电信", "联通", "移动", "教育网", "境外"},
		enumStrings(tencentLineValues.AnyOf[0]))
	require.Equal(t, JSONTypeString, tencentLineValues.AnyOf[1].Type)
	require.Equal(t, 1, *tencentLineValues.AnyOf[1].MinLength)
	require.Equal(t, JSONTypeArray, tencentLines.AnyOf[1].Type)
	require.Equal(t, 1, *tencentLines.AnyOf[1].Items.MinLength)
	require.Equal(t, tencentLineValues.AnyOf[0].Enum, tencentLines.AnyOf[1].Items.AnyOf[0].Enum)
}

func TestGeneratedSchemaRequiresNonOptionalTopLevelFields(t *testing.T) {
	data, err := GenSchema()
	require.NoError(t, err)

	var schema struct {
		Type                 string   `json:"type"`
		AdditionalProperties bool     `json:"additionalProperties"`
		Required             []string `json:"required"`
		Properties           struct {
			Log struct {
				Properties struct {
					Level struct {
						Default string `json:"default"`
					} `json:"level"`
					Output struct {
						Default string `json:"default"`
					} `json:"output"`
				} `json:"properties"`
			} `json:"log"`
			Domains struct {
				Items struct {
					Properties struct {
						Interval struct {
							Default string `json:"default"`
						} `json:"interval"`
						Timeout struct {
							Default string `json:"default"`
						} `json:"timeout"`
					} `json:"properties"`
				} `json:"items"`
			} `json:"domains"`
		} `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(data, &schema))
	require.Equal(t, JSONTypeObject, schema.Type)
	require.False(t, schema.AdditionalProperties)
	require.Equal(t, []string{"datasources", "providers", "domains", "services"}, schema.Required)
	require.Equal(t, "info", schema.Properties.Log.Properties.Level.Default)
	require.Equal(t, "stdout", schema.Properties.Log.Properties.Output.Default)
	require.Equal(t, "30s", schema.Properties.Domains.Items.Properties.Interval.Default)
	require.Equal(t, "15s", schema.Properties.Domains.Items.Properties.Timeout.Default)
}

func findVariant(t *testing.T, schema *jsonschema.Schema, typ string) *jsonschema.Schema {
	t.Helper()
	require.NotNil(t, schema.Items)
	for _, variant := range schema.Items.AnyOf {
		typeSchema := variant.Properties["type"]
		if typeSchema != nil && typeSchema.Const != nil && fmt.Sprint(*typeSchema.Const) == typ {
			return variant
		}
	}
	t.Fatalf("variant %q not found", typ)
	return nil
}

func enumStrings(schema *jsonschema.Schema) []string {
	values := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		values = append(values, fmt.Sprint(value))
	}
	return values
}

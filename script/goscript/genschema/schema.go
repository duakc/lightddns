package genschema

import (
	"encoding/json"
	"reflect"
	"slices"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/script/goscript/pkg/jsonschema"
	ipserverservice "github.com/duakc/lightddns/services/ipserver"
	prometheusservice "github.com/duakc/lightddns/services/prometheus"

	"github.com/duakc/mt"
)

var (
	optionsTypeMappingTable     = make(map[reflect.Type]*jsonschema.Schema)
	optionsTypeMappingTableOnce sync.Once
)

func optionsTypeMapping() map[reflect.Type]*jsonschema.Schema {
	optionsTypeMappingTableOnce.Do(func() {
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Listable[string]]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Listable[badyaml.Prefix]]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.EnvironmentVariable]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Prefix]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPHeader]()] = &jsonschema.Schema{
			Type:                 JSONTypeObject,
			AdditionalProperties: listAble(JSONTypeString),
		}

		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPMethod]()] = httpMethod()
		setDefault(optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPMethod]()], "GET")

		optionsTypeMappingTable[reflect.TypeFor[badyaml.Duration]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.URL]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DomainName]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Regex]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.JQ]()] = singleType(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[badyaml.StringOrNumber]()] = stringOr(singleType(JSONTypeNumber))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.LogLevel]()] = enumSchema(
			JSONTypeString, "debug", "info", "warn", "warning", "error", "dpanic", "panic", "fatal")
		setDefault(optionsTypeMappingTable[reflect.TypeFor[badyaml.LogLevel]()], "info")

		optionsTypeMappingTable[reflect.TypeFor[dialerx.DialStrategy]()] = enumSchema(JSONTypeString, mt.Map(
			[]dialerx.DialStrategy{dialerx.DialOnlyIPv4, dialerx.DialOnlyIPv6, dialerx.DialPreferIPv4, dialerx.DialPreferIPv6},
			func(s dialerx.DialStrategy) any {
				return s.String()
			})...)
		setDefault(optionsTypeMappingTable[reflect.TypeFor[dialerx.DialStrategy]()],
			dialerx.DialPreferIPv6.String())

		dnsObjectSchema := mt.Must(jsonschema.For[options.DNSOption](&jsonschema.ForOptions{
			TypeSchemas: optionsTypeMappingTable,
		}))
		dnsObjectSchema.Properties["type"] = enumSchema(
			JSONTypeString, transports.TransportTypeSystem, transports.TransportTypeTLS)
		setDefault(dnsObjectSchema.Properties["type"], transports.TransportTypeSystem)
		dnsObjectSchema.AllOf = append(dnsObjectSchema.AllOf, &jsonschema.Schema{
			If: &jsonschema.Schema{
				Properties: map[string]*jsonschema.Schema{
					"enabled": constSchema(true),
					"type":    constSchema(transports.TransportTypeTLS),
				},
				Required: []string{"enabled", "type"},
			},
			Then: &jsonschema.Schema{
				Properties: map[string]*jsonschema.Schema{
					"server": {Type: JSONTypeString, MinLength: new(1)},
				},
				Required: []string{"server"},
			},
		})
		optionsTypeMappingTable[reflect.TypeFor[options.DNSOption]()] = &jsonschema.Schema{
			AnyOf: []*jsonschema.Schema{
				enumSchema(JSONTypeString, transports.TransportTypeSystem),
				{Type: JSONTypeString, Pattern: `^tls://.+$`},
				dnsObjectSchema,
			},
		}
		optionsTypeMappingTable[reflect.TypeFor[options.CommandOutput]()] = enumSchema(JSONTypeString, mt.Map(
			[]options.CommandOutput{
				options.CommandOutputNone, options.CommandOutputStdout,
				options.CommandOutputStderr, options.CommandOutputAll,
			},
			func(s options.CommandOutput) any {
				return s
			})...)
		setDefault(optionsTypeMappingTable[reflect.TypeFor[options.CommandOutput]()],
			options.CommandOutputNone)
	})
	return optionsTypeMappingTable
}

func GenSchema() ([]byte, error) {
	rootSchema := &jsonschema.Schema{
		Type:                 JSONTypeObject,
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
	}
	var (
		logTag    = lookupTagIn[options.Options, options.LogOption]()
		domainTag = lookupTagIn[options.Options, options.DomainOption]()

		providerTag   = lookupTagIn[options.Options, options.ProviderOption]()
		datasourceTag = lookupTagIn[options.Options, options.DatasourceOption]()
		servicesTag   = lookupTagIn[options.Options, options.ServiceOption]()
	)
	rootSchema.Properties = make(map[string]*jsonschema.Schema)
	rootSchema.Properties[logTag] = mustFor[options.LogOption]()
	setDefault(rootSchema.Properties[logTag].Properties["output"], "stdout")
	rootSchema.Properties[domainTag] = mustFor[[]options.DomainOption]()
	requireArray(rootSchema.Properties[domainTag], false)
	rootSchema.Properties[domainTag].Items.Properties["domain"].MinLength = new(1)
	setDefault(rootSchema.Properties[domainTag].Items.Properties["interval"],
		constpkg.DefaultDomainUpdateInterval.String())
	setDefault(rootSchema.Properties[domainTag].Items.Properties["timeout"],
		constpkg.DefaultDomainTimeout.String())

	rootSchema.Properties[providerTag] = providerSchema()
	rootSchema.Properties[datasourceTag] = datasourceSchema()
	rootSchema.Properties[servicesTag] = servicesSchema()
	rootSchema.Required = []string{datasourceTag, providerTag, domainTag, servicesTag}

	return rootSchema.MarshalJSON()
}

func providerSchema() *jsonschema.Schema {
	return schemaFromRegistry(adapter.ProviderRegister)
}

func datasourceSchema() *jsonschema.Schema {
	return schemaFromRegistry(adapter.DatasourceRegister)
}

func servicesSchema() *jsonschema.Schema {
	return schemaFromRegistry(adapter.ServiceRegistry)
}

type schemaRegistry interface {
	Types() []string
	CreateOption(typ string) (any, error)
}

func schemaFromRegistry(reg schemaRegistry) *jsonschema.Schema {
	const typeTag = "type"

	types := reg.Types()
	slices.Sort(types)

	rootSchema := &jsonschema.Schema{Type: JSONTypeArray, Items: &jsonschema.Schema{}}
	for _, typ := range types {
		opt, err := reg.CreateOption(typ)
		if err != nil {
			continue
		}
		optType := reflect.TypeOf(opt).Elem()
		variantSchema := mt.Must(jsonschema.ForType(optType, &jsonschema.ForOptions{
			TypeSchemas: optionsTypeMapping(),
		}))
		variantSchema.Properties[typeTag].Const = new(any(typ))
		customizeVariantSchema(typ, variantSchema)
		rootSchema.Items.AnyOf = append(rootSchema.Items.AnyOf, variantSchema)
	}
	return rootSchema
}

func customizeVariantSchema(typ string, schema *jsonschema.Schema) {
	switch typ {
	case constpkg.DatasourceTypeCommand:
		schema.Properties["cmd"].AnyOf[0].MinLength = new(1)
		schema.Properties["cmd"].AnyOf[1].MinItems = new(1)
		schema.Properties["capture"] = enumSchema(JSONTypeString,
			options.CommandOutputStdout, options.CommandOutputStderr, options.CommandOutputAll)
		setDefault(schema.Properties["capture"], options.CommandOutputStdout)
	case constpkg.DatasourceTypeHTTP:
		schema.Properties["url"].MinLength = new(1)
		schema.Properties["url"].Pattern = `^https?://.+$`
	case constpkg.DatasourceTypeNetlink:
		one := 1.0
		schema.Properties["ifName"].MinLength = new(1)
		schema.Properties["ifIndex"].Minimum = &one
		schema.AnyOf = []*jsonschema.Schema{
			{Required: []string{"ifName"}},
			{Required: []string{"ifIndex"}},
		}
	case constpkg.DatasourceGroupTypeSum, constpkg.DatasourceGroupTypeFailover:
		requireArray(schema.Properties["datasources"], true)
	case constpkg.DatasourceGroupFilter:
		requireArray(schema.Properties["datasources"], false)
		requireArray(schema.Properties["rules"], true)
	case constpkg.ProviderTypeCloudflare:
		schema.Properties["token"].MinLength = new(1)
	case constpkg.ProviderTypeAliyun:
		schema.Properties["accessKeyId"].MinLength = new(1)
		schema.Properties["accessKeySecret"].MinLength = new(1)
		setProviderLineCompletions(schema.Properties["lines"],
			"default", "telecom", "unicom", "mobile", "oversea", "edu", "drpeng")
	case constpkg.ProviderTypeTencentCloud:
		schema.Properties["secretId"].MinLength = new(1)
		schema.Properties["secretKey"].MinLength = new(1)
		setProviderLineCompletions(schema.Properties["lines"],
			"默认", "电信", "联通", "移动", "教育网", "境外")
	case constpkg.ServiceTypePrometheus:
		setDefault(schema.Properties["port"], prometheusservice.DefaultPort)
		setDefault(schema.Properties["path"], prometheusservice.DefaultPath)
	case constpkg.ServiceTypeIPServer:
		setDefault(schema.Properties["port"], ipserverservice.DefaultPort)
		setDefault(schema.Properties["path"], ipserverservice.DefaultPath)
	}
}

// setProviderLineCompletions adds the provider's commonly used line values to
// the schema while keeping arbitrary strings valid. Providers can expose more
// lines than this stable, documented subset, and those values must remain
// usable without waiting for a schema update.
func setProviderLineCompletions(schema *jsonschema.Schema, lines ...string) {
	if schema == nil {
		return
	}
	for _, variant := range schema.AnyOf {
		if variant == nil {
			continue
		}
		switch variant.Type {
		case JSONTypeString:
			setLineValueCompletions(variant, lines...)
		case JSONTypeArray:
			setLineValueCompletions(variant.Items, lines...)
		}
	}
	if schema.Items != nil {
		setLineValueCompletions(schema.Items, lines...)
	}
}

func setLineValueCompletions(schema *jsonschema.Schema, lines ...string) {
	if schema == nil {
		return
	}
	schema.MinLength = new(1)
	freeForm := singleType(JSONTypeString)
	freeForm.MinLength = new(1)
	schema.AnyOf = []*jsonschema.Schema{
		enumSchema(JSONTypeString, mt.Map(lines, func(line string) any { return line })...),
		freeForm,
	}
}

func setDefault(schema *jsonschema.Schema, value any) {
	schema.Default = json.RawMessage(mt.Must(json.Marshal(value)))
}

func constSchema(value any) *jsonschema.Schema {
	return &jsonschema.Schema{Const: &value}
}

func requireArray(schema *jsonschema.Schema, nonEmpty bool) {
	schema.Type = JSONTypeArray
	schema.Types = nil
	if nonEmpty {
		schema.MinItems = new(1)
	}
}

func mustFor[T any]() *jsonschema.Schema {
	option := &jsonschema.ForOptions{
		TypeSchemas: optionsTypeMapping(),
	}
	return mt.Must(jsonschema.For[T](option))
}

func lookupTagIn[T, I any]() string {
	parentType := reflect.TypeFor[T]()
	if parentType.Kind() != reflect.Struct {
		panic("not a struct")
	}
	childType := reflect.TypeFor[I]()

	for i := 0; i < parentType.NumField(); i++ {
		field := parentType.Field(i)
		if field.Type == childType {
			return jsonschema.FieldJSONInfo(field).Name
		}
		if kind := field.Type.Kind(); (kind == reflect.Array || kind == reflect.Slice) &&
			field.Type.Elem() == childType {
			return jsonschema.FieldJSONInfo(field).Name
		}
	}
	panic("not found")
}

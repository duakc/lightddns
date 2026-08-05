package genschema

import (
	"reflect"
	"slices"
	"sync"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/script/goscript/pkg/jsonschema"

	"github.com/duakc/mt"
)

func dualStack(item *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{AnyOf: []*jsonschema.Schema{
		item,
		{
			Type: JSONTypeObject,
			Properties: map[string]*jsonschema.Schema{
				"ipv4": item,
				"ipv6": item,
			},
		},
	}}
}

var (
	optionsTypeMappingTable     = make(map[reflect.Type]*jsonschema.Schema)
	optionsTypeMappingTableOnce sync.Once
)

func optionsTypeMapping() map[reflect.Type]*jsonschema.Schema {
	optionsTypeMappingTableOnce.Do(func() {
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Listable[string]]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.Listable[badyaml.Prefix]]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.EnvironmentVariable]()] = listAble(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPHeader]()] = listAble(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPMethod]()] = httpMethod()

		optionsTypeMappingTable[reflect.TypeFor[badyaml.Duration]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.URL]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DomainName]()] = singleType(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[badyaml.StringOrNumber]()] = stringOr(singleType(JSONTypeNumber))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DualStack[string]]()] = dualStack(singleType(JSONTypeString))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DualStack[badyaml.URL]]()] = dualStack(singleType(JSONTypeString))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DualStack[badyaml.Listable[string]]]()] = dualStack(listAble(JSONTypeString))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.LogLevel]()] = enumSchema(JSONTypeString, "debug", "info", "warn", "error", "panic", "fatal")
		optionsTypeMappingTable[reflect.TypeFor[badyaml.NotEmpty[string]]()] = singleType(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[dialerx.DialStrategy]()] = enumSchema(JSONTypeString, mt.Map(
			[]dialerx.DialStrategy{dialerx.DialOnlyIPv4, dialerx.DialOnlyIPv6, dialerx.DialPreferIPv4, dialerx.DialPreferIPv6},
			func(s dialerx.DialStrategy) any {
				return s.String()
			})...)

		optionsTypeMappingTable[reflect.TypeFor[options.DNSOption]()] = mt.Must(jsonschema.For[options.DNSOption](nil))
	})
	return optionsTypeMappingTable
}

func GenSchema() ([]byte, error) {
	rootSchema := new(jsonschema.Schema)
	var (
		logTag    = lookupTagIn[options.Options, options.LogOption]()
		domainTag = lookupTagIn[options.Options, options.DomainOption]()

		providerTag   = lookupTagIn[options.Options, options.ProviderOption]()
		datasourceTag = lookupTagIn[options.Options, options.DatasourceOption]()
		servicesTag   = lookupTagIn[options.Options, options.ServiceOption]()
	)
	rootSchema.Properties = make(map[string]*jsonschema.Schema)
	rootSchema.Properties[logTag] = mustFor[options.LogOption]()
	rootSchema.Properties[domainTag] = mustFor[[]options.DomainOption]()

	rootSchema.Properties[providerTag] = providerSchema()
	rootSchema.Properties[datasourceTag] = datasourceSchema()
	rootSchema.Properties[servicesTag] = servicesSchema()

	// rootSchema.Required = append(rootSchema.Required,
	//	domainTag, providerTag, datasourceTag)

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
		rootSchema.Items.AnyOf = append(rootSchema.Items.AnyOf, variantSchema)
	}
	return rootSchema
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

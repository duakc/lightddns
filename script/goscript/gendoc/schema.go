package gendoc

import (
	"reflect"
	"sync"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/script/goscript/gendoc/jsonschema"

	"github.com/duakc/mt"
)

var (
	optionsTypeMappingTable     = make(map[reflect.Type]*jsonschema.Schema)
	optionsTypeMappingTableOnce sync.Once
)

func optionsTypeMapping() map[reflect.Type]*jsonschema.Schema {
	optionsTypeMappingTableOnce.Do(func() {
		listableString := listAble(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[badyaml.Listable[string]]()] = listableString
		optionsTypeMappingTable[reflect.TypeFor[badyaml.EnvironmentVariable]()] = listableString
		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPHeader]()] = listableString

		optionsTypeMappingTable[reflect.TypeFor[badyaml.HTTPMethod]()] = httpMethod()

		optionsTypeMappingTable[reflect.TypeFor[badyaml.Duration]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.URL]()] = singleType(JSONTypeString)
		optionsTypeMappingTable[reflect.TypeFor[badyaml.DomainName]()] = singleType(JSONTypeString)

		optionsTypeMappingTable[reflect.TypeFor[badyaml.StringOrNumber]()] = stringOr(singleType(JSONTypeNumber))
		optionsTypeMappingTable[reflect.TypeFor[badyaml.StringOrObject[options.DualStack[string]]]()] = stringOr(mt.Must(jsonschema.For[options.DualStack[string]](nil)))

		optionsTypeMappingTable[reflect.TypeFor[dialerx.DialStrategy]()] = enumSchema(JSONTypeString, mt.Map(
			[]dialerx.DialStrategy{dialerx.DialOnlyIPv4, dialerx.DialOnlyIPv6, dialerx.DialPreferIPv4, dialerx.DialPreferIPv6},
			func(s dialerx.DialStrategy) any {
				return s.String()
			})...)

		optionsTypeMappingTable[reflect.TypeFor[options.DNSOption]()] = mt.Must(jsonschema.For[options.DNSOption](nil))
	})
	return optionsTypeMappingTable
}

func GenSchema(structs []*StructDocument) ([]byte, error) {
	rootSchema := new(jsonschema.Schema)
	var (
		logTag    = lookupTagIn[options.Options, options.LogOption]()
		domainTag = lookupTagIn[options.Options, options.DomainOption]()

		providerTag   = lookupTagIn[options.Options, options.ProviderOption]()
		datasourceTag = lookupTagIn[options.Options, options.DatasourceOption]()
	)
	rootSchema.Properties = make(map[string]*jsonschema.Schema)
	rootSchema.Properties[logTag] = mustFor[options.LogOption]()
	rootSchema.Properties[domainTag] = mustFor[[]options.DomainOption]()

	rootSchema.Properties[providerTag] = providerSchema()
	rootSchema.Properties[datasourceTag] = datasourceSchema()

	rootSchema.Required = append(rootSchema.Required, domainTag, providerTag, datasourceTag)

	return rootSchema.MarshalJSON()
}

func providerSchema() *jsonschema.Schema {
	rootSchema := &jsonschema.Schema{Type: JSONTypeArray, Items: &jsonschema.Schema{}}
	schemaAddVariant[options.CloudflareProviderOption](rootSchema)
	return rootSchema
}

func datasourceSchema() *jsonschema.Schema {
	rootSchema := &jsonschema.Schema{Type: JSONTypeArray, Items: &jsonschema.Schema{}}
	schemaAddVariant[options.CommandDatasourceOption](rootSchema)
	schemaAddVariant[options.NetlinkDatasourceOption](rootSchema)
	schemaAddVariant[options.HTTPDatasourceOption](rootSchema)
	schemaAddVariant[options.DatasourceGroupSumOption](rootSchema)
	schemaAddVariant[options.DatasourceGroupFailoverOption](rootSchema)
	return rootSchema
}

func schemaAddVariant[T options.VariantOption](rootSchema *jsonschema.Schema) {
	const typeTag = "type"

	variantOptionSchema := mustFor[T]()
	variantOptionSchema.Properties[typeTag].Const = new(any(mt.Zero[T]().UsedType()))
	rootSchema.Items.AnyOf = append(rootSchema.Items.AnyOf, variantOptionSchema)
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

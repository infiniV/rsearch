package testdata

import "github.com/infiniv/rsearch/internal/schema"

// BenchmarkQueries contains various query patterns for benchmarking
var BenchmarkQueries = struct {
	Simple  []string
	Complex []string
	Long    []string
	Nested  []string
}{
	Simple: []string{
		"productCode:13w42",
		"status:active",
		"region:ca",
		"price:100",
		"productName:test",
	},
	Complex: []string{
		"productCode:13w42 AND region:ca AND status:active",
		"(region:ca OR region:ny) AND status:active AND price:[100 TO 500]",
		"productName:test* AND NOT status:discontinued",
		"price:[100 TO *] AND region:(ca OR ny OR tx) AND category:electronics",
		"status:active AND (region:ca OR region:ny) AND price:>=100 AND productCode:abc*",
		"productName:/test.*/ AND price:[50 TO 500} AND status:active OR discontinued",
		"_exists_:productCode AND region:ca AND price:>=100 AND price:<=1000",
		"productCode:test~2 AND region:ca AND status:active",
	},
	Long: []string{
		// 100+ terms with implicit OR
		"term1 term2 term3 term4 term5 term6 term7 term8 term9 term10 " +
			"term11 term12 term13 term14 term15 term16 term17 term18 term19 term20 " +
			"term21 term22 term23 term24 term25 term26 term27 term28 term29 term30 " +
			"term31 term32 term33 term34 term35 term36 term37 term38 term39 term40 " +
			"term41 term42 term43 term44 term45 term46 term47 term48 term49 term50 " +
			"term51 term52 term53 term54 term55 term56 term57 term58 term59 term60 " +
			"term61 term62 term63 term64 term65 term66 term67 term68 term69 term70 " +
			"term71 term72 term73 term74 term75 term76 term77 term78 term79 term80 " +
			"term81 term82 term83 term84 term85 term86 term87 term88 term89 term90 " +
			"term91 term92 term93 term94 term95 term96 term97 term98 term99 term100",
		// Long field query with many OR clauses
		"region:ca OR region:ny OR region:tx OR region:fl OR region:wa OR region:or OR region:nv OR region:az " +
			"OR region:co OR region:ut OR region:nm OR region:id OR region:mt OR region:wy OR region:nd OR region:sd " +
			"OR region:ne OR region:ks OR region:ok OR region:ar OR region:la OR region:ms OR region:al OR region:tn " +
			"OR region:ky OR region:wv OR region:va OR region:nc OR region:sc OR region:ga",
	},
	Nested: []string{
		"(((region:ca)))",
		"((region:ca OR region:ny) AND (status:active OR status:pending))",
		"(((region:ca OR region:ny) AND status:active) OR ((region:tx OR region:fl) AND status:pending))",
		"((((productCode:13w42 OR productCode:abc123) AND region:ca) OR (productCode:xyz789 AND region:ny)) AND status:active)",
		"(((((status:active)))))",
		"((region:ca AND status:active) OR (region:ny AND status:pending) OR (region:tx AND status:inactive))",
		"(((productCode:a AND region:b) OR (productCode:c AND region:d)) AND ((status:e OR status:f) AND category:g))",
		"((((((region:ca OR region:ny) AND status:active) OR region:tx) AND price:>=100) OR category:electronics) AND productCode:abc*)",
	},
}

// GetBenchmarkSchema returns a schema configured for benchmarks
func GetBenchmarkSchema() *schema.Schema {
	fields := map[string]schema.Field{
		"productCode": {Type: schema.TypeText, Indexed: true},
		"productName": {Type: schema.TypeText, Indexed: true},
		"region":      {Type: schema.TypeText, Indexed: true},
		"status":      {Type: schema.TypeText, Indexed: true},
		"category":    {Type: schema.TypeText, Indexed: true},
		"price":       {Type: schema.TypeFloat, Indexed: true},
		"quantity":    {Type: schema.TypeInteger, Indexed: true},
		"rodLength":   {Type: schema.TypeInteger},
		"active":      {Type: schema.TypeBoolean},
		"createdAt":   {Type: schema.TypeDateTime},
		"metadata":    {Type: schema.TypeJSON},
	}

	options := schema.SchemaOptions{
		NamingConvention: "snake_case",
		StrictOperators:  false,
		StrictFieldNames: false,
		DefaultField:     "productName",
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy:     true,
			Proximity: true,
			Regex:     true,
		},
	}

	return schema.NewSchema("benchmark", fields, options)
}

// GetSimpleSchema returns a minimal schema for simple benchmarks
func GetSimpleSchema() *schema.Schema {
	fields := map[string]schema.Field{
		"field1": {Type: schema.TypeText},
		"field2": {Type: schema.TypeInteger},
		"field3": {Type: schema.TypeText},
	}

	options := schema.SchemaOptions{
		NamingConvention: "none",
		DefaultField:     "field1",
	}

	return schema.NewSchema("simple", fields, options)
}

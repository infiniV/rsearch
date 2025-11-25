package main

import (
	"encoding/json"
	"fmt"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
)

func main() {
	mongoTranslator := translator.NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"price":        {Type: schema.TypeFloat},
		"status":       {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "product_code",
	})

	examples := []struct {
		name  string
		query string
	}{
		{"Simple field query", "product_code:13w42"},
		{"Boolean AND", "product_code:13w42 AND region:ca"},
		{"Boolean OR", "region:ca OR region:us"},
		{"Range inclusive", "price:[50 TO 500]"},
		{"Range exclusive", "price:{50 TO 500}"},
		{"Wildcard", "product_code:13*"},
		{"NOT operator", "NOT status:inactive"},
		{"Complex nested", "(status:active OR status:pending) AND region:ca"},
	}

	for _, ex := range examples {
		fmt.Printf("\n=== %s ===\n", ex.name)
		fmt.Printf("Query: %s\n", ex.query)

		p := parser.NewParser(ex.query)
		ast, err := p.Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		output, err := mongoTranslator.Translate(ast, testSchema)
		if err != nil {
			fmt.Printf("Translation error: %v\n", err)
			continue
		}

		filterJSON, _ := json.MarshalIndent(output.Filter, "", "  ")
		fmt.Printf("MongoDB Filter:\n%s\n", string(filterJSON))
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
)

func main() {
	fmt.Println("rsearch Translator Demo")
	fmt.Println("========================")
	fmt.Println()

	// Create schema registry
	schemaRegistry := schema.NewRegistry()

	// Create a test schema
	productsSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"rod_length":   {Type: schema.TypeInteger},
		"price":        {Type: schema.TypeFloat},
		"status":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	if err := schemaRegistry.Register(productsSchema); err != nil {
		log.Fatal(err)
	}

	// Create translator registry
	translatorRegistry := translator.NewRegistry()

	// Register PostgreSQL translator
	if err := translatorRegistry.Register("postgres", translator.NewPostgresTranslator()); err != nil {
		log.Fatal(err)
	}

	// Get translator
	postgres, err := translatorRegistry.Get("postgres")
	if err != nil {
		log.Fatal(err)
	}

	// Example translations using stub AST
	examples := []struct {
		description string
		ast         parser.Node
	}{
		{
			description: "Simple field query: productCode:13w42",
			ast: &parser.FieldQuery{
				Field: "product_code",
				Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
			},
		},
		{
			description: "Boolean AND: productCode:13w42 AND region:ca",
			ast: &parser.BinaryOp{
				Op: "AND",
				Left: &parser.FieldQuery{
					Field: "product_code",
					Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
				},
				Right: &parser.FieldQuery{
					Field: "region",
					Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
				},
			},
		},
		{
			description: "Range query: rodLength:[50 TO 500]",
			ast: &parser.RangeQuery{
				Field:          "rod_length",
				Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
				End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
				InclusiveStart: true,
				InclusiveEnd:   true,
			},
		},
		{
			description: "Complex nested: (productCode:13w42 AND region:ca) OR status:active",
			ast: &parser.BinaryOp{
				Op: "OR",
				Left: &parser.BinaryOp{
					Op: "AND",
					Left: &parser.FieldQuery{
						Field: "product_code",
						Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
					},
					Right: &parser.FieldQuery{
						Field: "region",
						Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
					},
				},
				Right: &parser.FieldQuery{
					Field: "status",
					Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
				},
			},
		},
	}

	// Translate each example
	for _, example := range examples {
		fmt.Printf("Query: %s\n", example.description)

		output, err := postgres.Translate(example.ast, productsSchema)
		if err != nil {
			fmt.Printf("  Error: %s\n\n", err)
			continue
		}

		fmt.Printf("  SQL:   %s\n", output.WhereClause)
		paramsJSON, _ := json.Marshal(output.Parameters)
		fmt.Printf("  Params: %s\n", paramsJSON)
		typesJSON, _ := json.Marshal(output.ParameterTypes)
		fmt.Printf("  Types: %s\n", typesJSON)
		fmt.Println()
	}

	fmt.Println("Note: Parser integration pending. These examples use stub AST nodes.")
	fmt.Println("After parser integration, queries will be parsed from strings automatically.")
}

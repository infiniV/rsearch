package main

import (
	"encoding/json"
	"fmt"
	"log"

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
	productsSchema := &schema.Schema{
		Name: "products",
		Fields: map[string]*schema.Field{
			"product_code": {Name: "product_code", Type: "text", Searchable: true},
			"region":       {Name: "region", Type: "text", Searchable: true},
			"rod_length":   {Name: "rod_length", Type: "number", Searchable: true},
			"price":        {Name: "price", Type: "number", Searchable: true},
			"status":       {Name: "status", Type: "text", Searchable: true},
		},
	}

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
		ast         translator.Node
	}{
		{
			description: "Simple field query: productCode:13w42",
			ast: &translator.FieldQuery{
				Field: "product_code",
				Value: "13w42",
			},
		},
		{
			description: "Boolean AND: productCode:13w42 AND region:ca",
			ast: &translator.BinaryOp{
				Op: "AND",
				Left: &translator.FieldQuery{
					Field: "product_code",
					Value: "13w42",
				},
				Right: &translator.FieldQuery{
					Field: "region",
					Value: "ca",
				},
			},
		},
		{
			description: "Range query: rodLength:[50 TO 500]",
			ast: &translator.RangeQuery{
				Field:          "rod_length",
				Start:          50,
				End:            500,
				InclusiveStart: true,
				InclusiveEnd:   true,
			},
		},
		{
			description: "Complex nested: (productCode:13w42 AND region:ca) OR status:active",
			ast: &translator.BinaryOp{
				Op: "OR",
				Left: &translator.BinaryOp{
					Op: "AND",
					Left: &translator.FieldQuery{
						Field: "product_code",
						Value: "13w42",
					},
					Right: &translator.FieldQuery{
						Field: "region",
						Value: "ca",
					},
				},
				Right: &translator.FieldQuery{
					Field: "status",
					Value: "active",
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

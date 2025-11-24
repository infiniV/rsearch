package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type TestCase struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Query       string   `json:"query"`
	Schema      string   `json:"schema"`
	Expected    Expected `json:"expected"`
}

type Expected struct {
	SQL            string                 `json:"sql"`
	Parameters     []interface{}          `json:"parameters"`
	ParameterTypes []string               `json:"parameterTypes"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

func main() {
	// Load test cases
	data, err := os.ReadFile("tests/testcases.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading test cases: %v\n", err)
		os.Exit(1)
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing test cases: %v\n", err)
		os.Exit(1)
	}

	// Group by category
	categories := make(map[string][]TestCase)
	for _, tc := range cases {
		categories[tc.Category] = append(categories[tc.Category], tc)
	}

	// Sort categories
	var categoryNames []string
	for name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	// Generate documentation
	doc := generateDocs(categoryNames, categories)

	// Write to file
	if err := os.WriteFile("docs/syntax-reference.md", []byte(doc), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing documentation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Documentation generated: docs/syntax-reference.md")
}

func generateDocs(categoryNames []string, categories map[string][]TestCase) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# rsearch Query Syntax Reference\n\n")
	sb.WriteString(fmt.Sprintf("*Auto-generated from test suite - Last updated: %s*\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString("This reference documents all supported query syntax patterns in rsearch, ")
	sb.WriteString("with examples showing the OpenSearch query and the resulting PostgreSQL translation.\n\n")

	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	for _, category := range categoryNames {
		anchor := strings.ToLower(strings.ReplaceAll(category, " ", "-"))
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", category, anchor))
	}
	sb.WriteString("\n---\n\n")

	// Categories
	for _, category := range categoryNames {
		sb.WriteString(fmt.Sprintf("## %s\n\n", category))

		for _, tc := range categories[category] {
			sb.WriteString(fmt.Sprintf("### %s\n\n", tc.Description))
			sb.WriteString(fmt.Sprintf("**Query:**\n```\n%s\n```\n\n", tc.Query))
			sb.WriteString(fmt.Sprintf("**PostgreSQL Translation:**\n```sql\n%s\n```\n\n", tc.Expected.SQL))

			if len(tc.Expected.Parameters) > 0 {
				sb.WriteString("**Parameters:**\n```json\n")
				params, _ := json.MarshalIndent(tc.Expected.Parameters, "", "  ")
				sb.WriteString(string(params))
				sb.WriteString("\n```\n\n")

				sb.WriteString("**Parameter Types:**\n```json\n")
				types, _ := json.MarshalIndent(tc.Expected.ParameterTypes, "", "  ")
				sb.WriteString(string(types))
				sb.WriteString("\n```\n\n")
			}

			if tc.Expected.Metadata != nil {
				sb.WriteString("**Metadata:**\n```json\n")
				metadata, _ := json.MarshalIndent(tc.Expected.Metadata, "", "  ")
				sb.WriteString(string(metadata))
				sb.WriteString("\n```\n\n")
			}

			sb.WriteString("---\n\n")
		}
	}

	// Footer
	sb.WriteString("## Additional Information\n\n")
	sb.WriteString("### Operator Precedence\n\n")
	sb.WriteString("From highest to lowest:\n")
	sb.WriteString("1. Boost (`^`)\n")
	sb.WriteString("2. Field queries (`:`)\n")
	sb.WriteString("3. Required/Prohibited (`+`, `-`)\n")
	sb.WriteString("4. NOT (`NOT`, `!`)\n")
	sb.WriteString("5. AND (`AND`, `&&`)\n")
	sb.WriteString("6. OR (`OR`, `||`)\n\n")

	sb.WriteString("### Operator Normalization\n\n")
	sb.WriteString("The parser normalizes operators to keyword form:\n")
	sb.WriteString("- `&&` → `AND`\n")
	sb.WriteString("- `||` → `OR`\n")
	sb.WriteString("- `!` → `NOT`\n\n")

	sb.WriteString("### Parameter Safety\n\n")
	sb.WriteString("All queries use parameterized statements with numbered placeholders (`$1`, `$2`, etc.) ")
	sb.WriteString("to prevent SQL injection attacks. Parameters are typed according to the schema field definitions.\n\n")

	sb.WriteString("### Schema Configuration\n\n")
	sb.WriteString("Field names are resolved according to the schema's naming convention (e.g., `snake_case`). ")
	sb.WriteString("Queries use the friendly field names (e.g., `productCode`) which are automatically ")
	sb.WriteString("mapped to database columns (e.g., `product_code`).\n\n")

	return sb.String()
}

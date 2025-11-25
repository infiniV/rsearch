package main

import (
	"strings"
	"testing"
)

func TestGenerateDocs(t *testing.T) {
	categories := []string{"Basic Queries", "Boolean Operators"}
	categoriesMap := map[string][]TestCase{
		"Basic Queries": {
			{
				Category:    "Basic Queries",
				Description: "Simple field query",
				Query:       "field:value",
				Schema:      "test",
				Expected: Expected{
					SQL:            "field = $1",
					Parameters:     []interface{}{"value"},
					ParameterTypes: []string{"text"},
				},
			},
		},
		"Boolean Operators": {
			{
				Category:    "Boolean Operators",
				Description: "AND operator",
				Query:       "a:1 AND b:2",
				Schema:      "test",
				Expected: Expected{
					SQL:            "(a = $1 AND b = $2)",
					Parameters:     []interface{}{"1", "2"},
					ParameterTypes: []string{"text", "text"},
				},
			},
		},
	}

	doc := generateDocs(categories, categoriesMap)

	// Verify header
	if !strings.Contains(doc, "# rsearch Query Syntax Reference") {
		t.Error("Missing main header")
	}

	// Verify table of contents
	if !strings.Contains(doc, "## Table of Contents") {
		t.Error("Missing table of contents")
	}
	if !strings.Contains(doc, "[Basic Queries](#basic-queries)") {
		t.Error("Missing Basic Queries in TOC")
	}
	if !strings.Contains(doc, "[Boolean Operators](#boolean-operators)") {
		t.Error("Missing Boolean Operators in TOC")
	}

	// Verify categories are present
	if !strings.Contains(doc, "## Basic Queries") {
		t.Error("Missing Basic Queries section")
	}
	if !strings.Contains(doc, "## Boolean Operators") {
		t.Error("Missing Boolean Operators section")
	}

	// Verify test cases are included
	if !strings.Contains(doc, "### Simple field query") {
		t.Error("Missing Simple field query description")
	}
	if !strings.Contains(doc, "field:value") {
		t.Error("Missing query example")
	}
	if !strings.Contains(doc, "field = $1") {
		t.Error("Missing SQL translation")
	}

	// Verify parameters section
	if !strings.Contains(doc, "**Parameters:**") {
		t.Error("Missing Parameters section")
	}
	if !strings.Contains(doc, "**Parameter Types:**") {
		t.Error("Missing Parameter Types section")
	}

	// Verify footer sections
	if !strings.Contains(doc, "## Additional Information") {
		t.Error("Missing Additional Information section")
	}
	if !strings.Contains(doc, "### Operator Precedence") {
		t.Error("Missing Operator Precedence section")
	}
	if !strings.Contains(doc, "### Operator Normalization") {
		t.Error("Missing Operator Normalization section")
	}
	if !strings.Contains(doc, "### Parameter Safety") {
		t.Error("Missing Parameter Safety section")
	}
	if !strings.Contains(doc, "### Schema Configuration") {
		t.Error("Missing Schema Configuration section")
	}
}

func TestGenerateDocsWithMetadata(t *testing.T) {
	categories := []string{"Boost"}
	categoriesMap := map[string][]TestCase{
		"Boost": {
			{
				Category:    "Boost",
				Description: "Boost query",
				Query:       "field:value^2",
				Schema:      "test",
				Expected: Expected{
					SQL:            "field = $1",
					Parameters:     []interface{}{"value"},
					ParameterTypes: []string{"text"},
					Metadata: map[string]interface{}{
						"boost": float64(2),
					},
				},
			},
		},
	}

	doc := generateDocs(categories, categoriesMap)

	// Verify metadata section is present
	if !strings.Contains(doc, "**Metadata:**") {
		t.Error("Missing Metadata section for boosted query")
	}
	if !strings.Contains(doc, "boost") {
		t.Error("Missing boost in metadata")
	}
}

func TestGenerateDocsEmptyCategories(t *testing.T) {
	categories := []string{}
	categoriesMap := map[string][]TestCase{}

	doc := generateDocs(categories, categoriesMap)

	// Should still have header and footer
	if !strings.Contains(doc, "# rsearch Query Syntax Reference") {
		t.Error("Missing main header")
	}
	if !strings.Contains(doc, "## Additional Information") {
		t.Error("Missing Additional Information section")
	}
}

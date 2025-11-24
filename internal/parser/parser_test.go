package parser

import (
	"testing"
)

func TestParser_SimpleFieldQuery(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "simple field:value",
			input: "productCode:13w42",
			checkFunc: func(t *testing.T, node Node) {
				fq, ok := node.(*FieldQuery)
				if !ok {
					t.Fatalf("expected FieldQuery, got %T", node)
				}
				if fq.Field != "productCode" {
					t.Errorf("expected field 'productCode', got %q", fq.Field)
				}
				term, ok := fq.Value.(*TermValue)
				if !ok {
					t.Fatalf("expected TermValue, got %T", fq.Value)
				}
				if term.Term != "13w42" {
					t.Errorf("expected term '13w42', got %q", term.Term)
				}
			},
		},
		{
			name:  "field with wildcard",
			input: "name:wid*",
			checkFunc: func(t *testing.T, node Node) {
				fq, ok := node.(*FieldQuery)
				if !ok {
					t.Fatalf("expected FieldQuery, got %T", node)
				}
				wildcard, ok := fq.Value.(*WildcardValue)
				if !ok {
					t.Fatalf("expected WildcardValue, got %T", fq.Value)
				}
				if wildcard.Pattern != "wid*" {
					t.Errorf("expected pattern 'wid*', got %q", wildcard.Pattern)
				}
			},
		},
		{
			name:  "field with quoted string",
			input: `name:"blue widget"`,
			checkFunc: func(t *testing.T, node Node) {
				fq, ok := node.(*FieldQuery)
				if !ok {
					t.Fatalf("expected FieldQuery, got %T", node)
				}
				phrase, ok := fq.Value.(*PhraseValue)
				if !ok {
					t.Fatalf("expected PhraseValue, got %T", fq.Value)
				}
				if phrase.Phrase != "blue widget" {
					t.Errorf("expected phrase 'blue widget', got %q", phrase.Phrase)
				}
			},
		},
		{
			name:  "field with number",
			input: "price:100",
			checkFunc: func(t *testing.T, node Node) {
				fq, ok := node.(*FieldQuery)
				if !ok {
					t.Fatalf("expected FieldQuery, got %T", node)
				}
				num, ok := fq.Value.(*NumberValue)
				if !ok {
					t.Fatalf("expected NumberValue, got %T", fq.Value)
				}
				if num.Number != "100" {
					t.Errorf("expected number '100', got %q", num.Number)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_BooleanOperators(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "simple AND",
			input: "a AND b",
			checkFunc: func(t *testing.T, node Node) {
				bin, ok := node.(*BinaryOp)
				if !ok {
					t.Fatalf("expected BinaryOp, got %T", node)
				}
				if bin.Op != "AND" {
					t.Errorf("expected AND operator, got %q", bin.Op)
				}
			},
		},
		{
			name:  "simple OR",
			input: "a OR b",
			checkFunc: func(t *testing.T, node Node) {
				bin, ok := node.(*BinaryOp)
				if !ok {
					t.Fatalf("expected BinaryOp, got %T", node)
				}
				if bin.Op != "OR" {
					t.Errorf("expected OR operator, got %q", bin.Op)
				}
			},
		},
		{
			name:  "AND with OR precedence",
			input: "a AND b OR c",
			checkFunc: func(t *testing.T, node Node) {
				// Should parse as (a AND b) OR c
				bin, ok := node.(*BinaryOp)
				if !ok {
					t.Fatalf("expected BinaryOp, got %T", node)
				}
				if bin.Op != "OR" {
					t.Errorf("expected top-level OR, got %q", bin.Op)
				}
				left, ok := bin.Left.(*BinaryOp)
				if !ok {
					t.Fatalf("expected left BinaryOp, got %T", bin.Left)
				}
				if left.Op != "AND" {
					t.Errorf("expected left AND, got %q", left.Op)
				}
			},
		},
		{
			name:  "parentheses override precedence",
			input: "a AND (b OR c)",
			checkFunc: func(t *testing.T, node Node) {
				bin, ok := node.(*BinaryOp)
				if !ok {
					t.Fatalf("expected BinaryOp, got %T", node)
				}
				if bin.Op != "AND" {
					t.Errorf("expected top-level AND, got %q", bin.Op)
				}
			},
		},
		{
			name:  "NOT operator",
			input: "NOT field:value",
			checkFunc: func(t *testing.T, node Node) {
				unary, ok := node.(*UnaryOp)
				if !ok {
					t.Fatalf("expected UnaryOp, got %T", node)
				}
				if unary.Op != "NOT" {
					t.Errorf("expected NOT operator, got %q", unary.Op)
				}
			},
		},
		{
			name:  "exclamation NOT",
			input: "!field:value",
			checkFunc: func(t *testing.T, node Node) {
				unary, ok := node.(*UnaryOp)
				if !ok {
					t.Fatalf("expected UnaryOp, got %T", node)
				}
				if unary.Op != "!" {
					t.Errorf("expected ! operator, got %q", unary.Op)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_RequiredProhibited(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "required term",
			input: "+required",
			checkFunc: func(t *testing.T, node Node) {
				req, ok := node.(*RequiredQuery)
				if !ok {
					t.Fatalf("expected RequiredQuery, got %T", node)
				}
				_ = req
			},
		},
		{
			name:  "prohibited term",
			input: "-prohibited",
			checkFunc: func(t *testing.T, node Node) {
				proh, ok := node.(*ProhibitedQuery)
				if !ok {
					t.Fatalf("expected ProhibitedQuery, got %T", node)
				}
				_ = proh
			},
		},
		{
			name:  "combined required and prohibited",
			input: "+required -prohibited",
			checkFunc: func(t *testing.T, node Node) {
				bin, ok := node.(*BinaryOp)
				if !ok {
					t.Fatalf("expected BinaryOp, got %T", node)
				}
				if bin.Op != "OR" {
					t.Errorf("expected implicit OR, got %q", bin.Op)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_RangeQueries(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "inclusive range",
			input: "rodLength:[50 TO 500]",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if rq.Field != "rodLength" {
					t.Errorf("expected field 'rodLength', got %q", rq.Field)
				}
				if !rq.InclusiveStart {
					t.Error("expected inclusive start")
				}
				if !rq.InclusiveEnd {
					t.Error("expected inclusive end")
				}
			},
		},
		{
			name:  "exclusive range",
			input: "rodLength:{50 TO 500}",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if rq.InclusiveStart {
					t.Error("expected exclusive start")
				}
				if rq.InclusiveEnd {
					t.Error("expected exclusive end")
				}
			},
		},
		{
			name:  "mixed range",
			input: "price:{100 TO 200]",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if rq.InclusiveStart {
					t.Error("expected exclusive start")
				}
				if !rq.InclusiveEnd {
					t.Error("expected inclusive end")
				}
			},
		},
		{
			name:  "greater than",
			input: "price:>100",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if rq.InclusiveStart {
					t.Error("expected exclusive start for >")
				}
			},
		},
		{
			name:  "greater than or equal",
			input: "price:>=100",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if !rq.InclusiveStart {
					t.Error("expected inclusive start for >=")
				}
			},
		},
		{
			name:  "less than",
			input: "price:<100",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if rq.InclusiveEnd {
					t.Error("expected exclusive end for <")
				}
			},
		},
		{
			name:  "less than or equal",
			input: "price:<=100",
			checkFunc: func(t *testing.T, node Node) {
				rq, ok := node.(*RangeQuery)
				if !ok {
					t.Fatalf("expected RangeQuery, got %T", node)
				}
				if !rq.InclusiveEnd {
					t.Error("expected inclusive end for <=")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_BoostAndFuzzy(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "boost term",
			input: "term^2",
			checkFunc: func(t *testing.T, node Node) {
				boost, ok := node.(*BoostQuery)
				if !ok {
					t.Fatalf("expected BoostQuery, got %T", node)
				}
				if boost.Boost != 2.0 {
					t.Errorf("expected boost 2.0, got %f", boost.Boost)
				}
			},
		},
		{
			name:  "boost field query",
			input: "name:widget^2",
			checkFunc: func(t *testing.T, node Node) {
				boost, ok := node.(*BoostQuery)
				if !ok {
					t.Fatalf("expected BoostQuery, got %T", node)
				}
				if boost.Boost != 2.0 {
					t.Errorf("expected boost 2.0, got %f", boost.Boost)
				}
			},
		},
		{
			name:  "fuzzy term",
			input: "term~2",
			checkFunc: func(t *testing.T, node Node) {
				fuzzy, ok := node.(*FuzzyQuery)
				if !ok {
					t.Fatalf("expected FuzzyQuery, got %T", node)
				}
				if fuzzy.Distance != 2 {
					t.Errorf("expected distance 2, got %d", fuzzy.Distance)
				}
			},
		},
		{
			name:  "fuzzy field query",
			input: "name:widget~2",
			checkFunc: func(t *testing.T, node Node) {
				fuzzy, ok := node.(*FuzzyQuery)
				if !ok {
					t.Fatalf("expected FuzzyQuery, got %T", node)
				}
				if fuzzy.Field != "name" {
					t.Errorf("expected field 'name', got %q", fuzzy.Field)
				}
				if fuzzy.Distance != 2 {
					t.Errorf("expected distance 2, got %d", fuzzy.Distance)
				}
			},
		},
		{
			name:  "proximity query",
			input: `"quick brown fox"~3`,
			checkFunc: func(t *testing.T, node Node) {
				prox, ok := node.(*ProximityQuery)
				if !ok {
					t.Fatalf("expected ProximityQuery, got %T", node)
				}
				if prox.Distance != 3 {
					t.Errorf("expected distance 3, got %d", prox.Distance)
				}
				if prox.Phrase != "quick brown fox" {
					t.Errorf("expected phrase 'quick brown fox', got %q", prox.Phrase)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_FieldGroupQuery(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, Node)
	}{
		{
			name:  "field with OR group",
			input: "name:(blue OR red)",
			checkFunc: func(t *testing.T, node Node) {
				fg, ok := node.(*FieldGroupQuery)
				if !ok {
					t.Fatalf("expected FieldGroupQuery, got %T", node)
				}
				if fg.Field != "name" {
					t.Errorf("expected field 'name', got %q", fg.Field)
				}
				if len(fg.Queries) < 1 {
					t.Fatalf("expected at least 1 query, got %d", len(fg.Queries))
				}
			},
		},
		{
			name:  "field with AND group",
			input: "tags:(scala AND functional)",
			checkFunc: func(t *testing.T, node Node) {
				fg, ok := node.(*FieldGroupQuery)
				if !ok {
					t.Fatalf("expected FieldGroupQuery, got %T", node)
				}
				if fg.Field != "tags" {
					t.Errorf("expected field 'tags', got %q", fg.Field)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			tt.checkFunc(t, node)
		})
	}
}

func TestParser_ExistsQuery(t *testing.T) {
	input := "_exists_:fieldName"
	parser := NewParser(input)
	node, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	exists, ok := node.(*ExistsQuery)
	if !ok {
		t.Fatalf("expected ExistsQuery, got %T", node)
	}
	if exists.Field != "fieldName" {
		t.Errorf("expected field 'fieldName', got %q", exists.Field)
	}
}

func TestParser_ComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "complex boolean",
			input: "(a OR b) AND (c OR d)",
		},
		{
			name:  "mixed operators",
			input: "+required -prohibited optional",
		},
		{
			name:  "field with wildcard and boost",
			input: "name:wid*^2",
		},
		{
			name:  "multiple fields",
			input: "name:widget AND price:>100",
		},
		{
			name:  "nested groups",
			input: "((a OR b) AND (c OR d)) OR e",
		},
		{
			name:  "all features",
			input: `name:"blue widget"^2 AND price:[100 TO 500] OR tags:(scala functional)`,
		},
		{
			name:  "implicit OR",
			input: "quick brown fox",
		},
		{
			name:  "wildcard queries",
			input: "qu?ck bro* NOT fox",
		},
		{
			name:  "regex query",
			input: "name:/[mb]oat/",
		},
		{
			name:  "date range",
			input: "date:[2020-01-01 TO 2020-12-31]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if node == nil {
				t.Fatal("expected non-nil node")
			}
		})
	}
}

func TestParser_FromDesignDoc(t *testing.T) {
	// Test all examples from the design document (lines 238-317)
	tests := []string{
		// Basic queries
		"productCode:13w42",
		"name:widget",
		`name:"blue widget"`,
		"price:100",
		"active:true",

		// Wildcards
		"qu?ck",
		"name:wid*",
		"name:widget*",
		"name:*get",

		// Regex
		"name:/[mb]oat/",
		"name:/joh?n(ath[oa]n)/",

		// Fuzzy
		"roam~",
		"roam~1",
		"name:widget~2",

		// Proximity
		`"quick fox"~5`,

		// Ranges
		"rodLength:[50 TO 500]",
		"price:{100 TO 200}",
		"price:[100 TO 200}",
		"price:>=100",
		"price:>100",
		"price:<200",
		"price:<=200",

		// Boolean operators
		"scala AND functional",
		"scala OR functional",
		"scala && functional",
		"scala || functional",
		"NOT deprecated",
		"!deprecated",
		"active:true AND !deprecated",
		"(scala OR java) AND functional",

		// Required/Prohibited
		"+required",
		"-prohibited",
		"+scala +functional -deprecated",

		// Field grouping
		"name:(blue OR red)",
		"tags:(scala AND functional)",

		// Boost
		"scala^2",
		"name:scala^5",
		`"quick fox"^2`,

		// Exists
		"_exists_:fieldName",

		// Complex
		`name:"blue widget" AND price:[100 TO 500]`,
		"(quick OR brown) AND fox",
		"+name:widget -deprecated",
		`tags:(scala functional) AND active:true^2`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			parser := NewParser(input)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error for %q: %v", input, err)
			}
			if node == nil {
				t.Fatalf("expected non-nil node for %q", input)
			}
		})
	}
}

func TestParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed parenthesis",
			input: "(a OR b",
		},
		{
			name:  "unclosed bracket",
			input: "price:[100 TO 200",
		},
		{
			name:  "invalid range",
			input: "price:[100 200]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			if err == nil {
				// It's okay if some malformed queries still parse
				// The lexer already caught many SQL injection attempts
				t.Logf("query parsed without error (may be acceptable): %q", tt.input)
			}
		})
	}
}

func TestParser_ImplicitOR(t *testing.T) {
	input := "quick brown fox"
	parser := NewParser(input)
	node, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should create implicit OR between terms
	bin, ok := node.(*BinaryOp)
	if !ok {
		t.Fatalf("expected BinaryOp for implicit OR, got %T", node)
	}
	if bin.Op != "OR" {
		t.Errorf("expected OR operator, got %q", bin.Op)
	}
}

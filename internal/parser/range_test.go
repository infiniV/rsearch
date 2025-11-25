package parser

import (
	"testing"
)

// TestRangeQueryParsing tests comprehensive range query parsing scenarios
func TestRangeQueryParsing(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		wantType           string
		wantField          string
		wantStart          string
		wantEnd            string
		wantInclusiveStart bool
		wantInclusiveEnd   bool
		wantError          bool
	}{
		{
			name:               "Inclusive both sides - bracket notation",
			input:              "age:[18 TO 65]",
			wantType:           "RangeQuery",
			wantField:          "age",
			wantStart:          "18",
			wantEnd:            "65",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Exclusive both sides - brace notation",
			input:              "price:{100 TO 1000}",
			wantType:           "RangeQuery",
			wantField:          "price",
			wantStart:          "100",
			wantEnd:            "1000",
			wantInclusiveStart: false,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Mixed - inclusive start, exclusive end",
			input:              "score:[50 TO 100}",
			wantType:           "RangeQuery",
			wantField:          "score",
			wantStart:          "50",
			wantEnd:            "100",
			wantInclusiveStart: true,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Mixed - exclusive start, inclusive end",
			input:              "rating:{0 TO 5]",
			wantType:           "RangeQuery",
			wantField:          "rating",
			wantStart:          "0",
			wantEnd:            "5",
			wantInclusiveStart: false,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Greater than or equal - comparison syntax",
			input:              "age:>=18",
			wantType:           "RangeQuery",
			wantField:          "age",
			wantStart:          "18",
			wantEnd:            "*",
			wantInclusiveStart: true,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Greater than - comparison syntax",
			input:              "price:>100",
			wantType:           "RangeQuery",
			wantField:          "price",
			wantStart:          "100",
			wantEnd:            "*",
			wantInclusiveStart: false,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Less than or equal - comparison syntax",
			input:              "age:<=65",
			wantType:           "RangeQuery",
			wantField:          "age",
			wantStart:          "*",
			wantEnd:            "65",
			wantInclusiveStart: false,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Less than - comparison syntax",
			input:              "score:<100",
			wantType:           "RangeQuery",
			wantField:          "score",
			wantStart:          "*",
			wantEnd:            "100",
			wantInclusiveStart: false,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Date range - inclusive",
			input:              "created:[2024-01-01 TO 2024-12-31]",
			wantType:           "RangeQuery",
			wantField:          "created",
			wantStart:          "2024-01-01",
			wantEnd:            "2024-12-31",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Date range - exclusive",
			input:              "updated:{2024-01-01 TO 2024-12-31}",
			wantType:           "RangeQuery",
			wantField:          "updated",
			wantStart:          "2024-01-01",
			wantEnd:            "2024-12-31",
			wantInclusiveStart: false,
			wantInclusiveEnd:   false,
			wantError:          false,
		},
		{
			name:               "Decimal range - inclusive",
			input:              "rating:[0.0 TO 5.0]",
			wantType:           "RangeQuery",
			wantField:          "rating",
			wantStart:          "0.0",
			wantEnd:            "5.0",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Negative numbers - inclusive",
			input:              `temperature:["-10" TO "40"]`,
			wantType:           "RangeQuery",
			wantField:          "temperature",
			wantStart:          "-10",
			wantEnd:            "40",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "String range - alphabetical",
			input:              "name:[alice TO zoe]",
			wantType:           "RangeQuery",
			wantField:          "name",
			wantStart:          "alice",
			wantEnd:            "zoe",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Quoted string range",
			input:              `title:["Alice in Wonderland" TO "Zoo Story"]`,
			wantType:           "RangeQuery",
			wantField:          "title",
			wantStart:          "Alice in Wonderland",
			wantEnd:            "Zoo Story",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Unbounded range - open start",
			input:              "age:[* TO 18]",
			wantType:           "RangeQuery",
			wantField:          "age",
			wantStart:          "*",
			wantEnd:            "18",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Unbounded range - open end",
			input:              "price:[100 TO *]",
			wantType:           "RangeQuery",
			wantField:          "price",
			wantStart:          "100",
			wantEnd:            "*",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Large numbers",
			input:              "population:[1000000 TO 10000000]",
			wantType:           "RangeQuery",
			wantField:          "population",
			wantStart:          "1000000",
			wantEnd:            "10000000",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Scientific notation",
			input:              "value:[1e6 TO 1e9]",
			wantType:           "RangeQuery",
			wantField:          "value",
			wantStart:          "1e6",
			wantEnd:            "1e9",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Timestamp range with microseconds",
			input:              `timestamp:["2024-01-01T00:00:00.000000" TO "2024-12-31T23:59:59.999999"]`,
			wantType:           "RangeQuery",
			wantField:          "timestamp",
			wantStart:          "2024-01-01T00:00:00.000000",
			wantEnd:            "2024-12-31T23:59:59.999999",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "IP address range",
			input:              `ip:["192.168.0.1" TO "192.168.0.255"]`,
			wantType:           "RangeQuery",
			wantField:          "ip",
			wantStart:          "192.168.0.1",
			wantEnd:            "192.168.0.255",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Zero value ranges",
			input:              "count:[0 TO 100]",
			wantType:           "RangeQuery",
			wantField:          "count",
			wantStart:          "0",
			wantEnd:            "100",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
		{
			name:               "Single character range",
			input:              "grade:[A TO F]",
			wantType:           "RangeQuery",
			wantField:          "grade",
			wantStart:          "A",
			wantEnd:            "F",
			wantInclusiveStart: true,
			wantInclusiveEnd:   true,
			wantError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			ast, err := parser.Parse()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if ast == nil {
				t.Fatal("AST is nil")
			}

			// Check node type
			rq, ok := ast.(*RangeQuery)
			if !ok {
				t.Fatalf("Expected RangeQuery, got %T", ast)
			}

			// Verify field
			if rq.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", rq.Field, tt.wantField)
			}

			// Verify start value
			startValue := rq.Start.Value()
			if startValue != tt.wantStart {
				t.Errorf("Start = %q, want %q", startValue, tt.wantStart)
			}

			// Verify end value
			endValue := rq.End.Value()
			if endValue != tt.wantEnd {
				t.Errorf("End = %q, want %q", endValue, tt.wantEnd)
			}

			// Verify inclusiveness
			if rq.InclusiveStart != tt.wantInclusiveStart {
				t.Errorf("InclusiveStart = %v, want %v", rq.InclusiveStart, tt.wantInclusiveStart)
			}

			if rq.InclusiveEnd != tt.wantInclusiveEnd {
				t.Errorf("InclusiveEnd = %v, want %v", rq.InclusiveEnd, tt.wantInclusiveEnd)
			}

			// Verify node type string
			if rq.Type() != tt.wantType {
				t.Errorf("Type() = %q, want %q", rq.Type(), tt.wantType)
			}
		})
	}
}

// TestRangeQueryInComplexQueries tests range queries combined with other query types
func TestRangeQueryInComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, Node)
	}{
		{
			name:  "Range AND field query",
			input: "age:[18 TO 65] AND status:active",
			check: func(t *testing.T, ast Node) {
				bo, ok := ast.(*BinaryOp)
				if !ok {
					t.Fatalf("Expected BinaryOp, got %T", ast)
				}
				if bo.Op != "AND" {
					t.Errorf("Op = %q, want AND", bo.Op)
				}

				// Check left is RangeQuery
				_, ok = bo.Left.(*RangeQuery)
				if !ok {
					t.Errorf("Left should be RangeQuery, got %T", bo.Left)
				}

				// Check right is FieldQuery
				_, ok = bo.Right.(*FieldQuery)
				if !ok {
					t.Errorf("Right should be FieldQuery, got %T", bo.Right)
				}
			},
		},
		{
			name:  "Range OR range query",
			input: "age:[18 TO 30] OR age:[60 TO 100]",
			check: func(t *testing.T, ast Node) {
				bo, ok := ast.(*BinaryOp)
				if !ok {
					t.Fatalf("Expected BinaryOp, got %T", ast)
				}
				if bo.Op != "OR" {
					t.Errorf("Op = %q, want OR", bo.Op)
				}

				// Both sides should be RangeQuery
				_, ok = bo.Left.(*RangeQuery)
				if !ok {
					t.Errorf("Left should be RangeQuery, got %T", bo.Left)
				}

				_, ok = bo.Right.(*RangeQuery)
				if !ok {
					t.Errorf("Right should be RangeQuery, got %T", bo.Right)
				}
			},
		},
		{
			name:  "Multiple range queries with AND",
			input: "age:[18 TO 65] AND salary:[50000 TO 150000] AND experience:[2 TO 10]",
			check: func(t *testing.T, ast Node) {
				// Should be nested BinaryOps
				_, ok := ast.(*BinaryOp)
				if !ok {
					t.Fatalf("Expected BinaryOp, got %T", ast)
				}

				// Walk the tree to ensure all leaf nodes are RangeQuery
				var walkAndCount func(Node) int
				walkAndCount = func(n Node) int {
					if _, ok := n.(*RangeQuery); ok {
						return 1
					}
					if b, ok := n.(*BinaryOp); ok {
						return walkAndCount(b.Left) + walkAndCount(b.Right)
					}
					return 0
				}

				count := walkAndCount(ast)
				if count != 3 {
					t.Errorf("Expected 3 RangeQuery nodes, got %d", count)
				}
			},
		},
		{
			name:  "Range with comparison operator >= combined",
			input: "price:>=100 AND category:electronics",
			check: func(t *testing.T, ast Node) {
				bo, ok := ast.(*BinaryOp)
				if !ok {
					t.Fatalf("Expected BinaryOp, got %T", ast)
				}

				// Left should be RangeQuery (>= is parsed as range)
				rq, ok := bo.Left.(*RangeQuery)
				if !ok {
					t.Errorf("Left should be RangeQuery, got %T", bo.Left)
				} else {
					if !rq.InclusiveStart {
						t.Error("Expected InclusiveStart to be true for >=")
					}
				}

				// Right should be FieldQuery
				_, ok = bo.Right.(*FieldQuery)
				if !ok {
					t.Errorf("Right should be FieldQuery, got %T", bo.Right)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			ast, err := parser.Parse()

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if ast == nil {
				t.Fatal("AST is nil")
			}

			tt.check(t, ast)
		})
	}
}

// TestRangeQueryEdgeCases tests edge cases and error conditions
func TestRangeQueryEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "Same start and end - inclusive",
			input:     "age:[18 TO 18]",
			wantError: false,
		},
		{
			name:      "Same start and end - exclusive",
			input:     "age:{18 TO 18}",
			wantError: false,
		},
		{
			name:      "Whitespace in range",
			input:     "age:[  18  TO  65  ]",
			wantError: false,
		},
		{
			name:      "Mixed quotes in range values",
			input:     `name:["Alice" TO "Bob"]`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			ast, err := parser.Parse()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if ast == nil {
				t.Error("AST is nil")
			}
		})
	}
}

package parser

import (
	"testing"
)

func TestLexer_BasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple field query",
			input:    "productCode:13w42",
			expected: []TokenType{STRING, COLON, STRING, EOF},
		},
		{
			name:     "field with wildcard",
			input:    "name:wid*",
			expected: []TokenType{STRING, COLON, WILDCARD, EOF},
		},
		{
			name:     "field with question wildcard",
			input:    "name:wi?get",
			expected: []TokenType{STRING, COLON, WILDCARD, EOF},
		},
		{
			name:     "quoted string",
			input:    `name:"blue widget"`,
			expected: []TokenType{STRING, COLON, QUOTED_STRING, EOF},
		},
		{
			name:     "number value",
			input:    "price:100",
			expected: []TokenType{STRING, COLON, NUMBER, EOF},
		},
		{
			name:     "decimal number",
			input:    "price:99.99",
			expected: []TokenType{STRING, COLON, NUMBER, EOF},
		},
		{
			name:     "AND operator",
			input:    "a AND b",
			expected: []TokenType{STRING, AND, STRING, EOF},
		},
		{
			name:     "OR operator",
			input:    "a OR b",
			expected: []TokenType{STRING, OR, STRING, EOF},
		},
		{
			name:     "NOT operator",
			input:    "NOT a",
			expected: []TokenType{NOT, STRING, EOF},
		},
		{
			name:     "double ampersand",
			input:    "a && b",
			expected: []TokenType{STRING, AND, STRING, EOF},
		},
		{
			name:     "double pipe",
			input:    "a || b",
			expected: []TokenType{STRING, OR, STRING, EOF},
		},
		{
			name:     "exclamation NOT",
			input:    "!a",
			expected: []TokenType{NOT, STRING, EOF},
		},
		{
			name:     "required term",
			input:    "+required",
			expected: []TokenType{PLUS, STRING, EOF},
		},
		{
			name:     "prohibited term",
			input:    "-prohibited",
			expected: []TokenType{MINUS, STRING, EOF},
		},
		{
			name:     "boost",
			input:    "name:widget^2",
			expected: []TokenType{STRING, COLON, STRING, CARET, NUMBER, EOF},
		},
		{
			name:     "fuzzy",
			input:    "name:widget~2",
			expected: []TokenType{STRING, COLON, STRING, TILDE, NUMBER, EOF},
		},
		{
			name:     "parentheses",
			input:    "(a OR b)",
			expected: []TokenType{LPAREN, STRING, OR, STRING, RPAREN, EOF},
		},
		{
			name:     "square brackets range",
			input:    "[50 TO 500]",
			expected: []TokenType{LBRACKET, NUMBER, TO, NUMBER, RBRACKET, EOF},
		},
		{
			name:     "curly braces range",
			input:    "{50 TO 500}",
			expected: []TokenType{LBRACE, NUMBER, TO, NUMBER, RBRACE, EOF},
		},
		{
			name:     "greater than",
			input:    "price:>100",
			expected: []TokenType{STRING, COLON, GT, NUMBER, EOF},
		},
		{
			name:     "greater than or equal",
			input:    "price:>=100",
			expected: []TokenType{STRING, COLON, GTE, NUMBER, EOF},
		},
		{
			name:     "less than",
			input:    "price:<100",
			expected: []TokenType{STRING, COLON, LT, NUMBER, EOF},
		},
		{
			name:     "less than or equal",
			input:    "price:<=100",
			expected: []TokenType{STRING, COLON, LTE, NUMBER, EOF},
		},
		{
			name:     "exists query",
			input:    "_exists_:field",
			expected: []TokenType{EXISTS, COLON, STRING, EOF},
		},
		{
			name:     "regex pattern",
			input:    "name:/wi[dg]get/",
			expected: []TokenType{STRING, COLON, REGEX, EOF},
		},
		{
			name:     "complex regex",
			input:    "name:/[a-z]+/",
			expected: []TokenType{STRING, COLON, REGEX, EOF},
		},
		{
			name:     "field with dots",
			input:    "user.name:john",
			expected: []TokenType{STRING, COLON, STRING, EOF},
		},
		{
			name:     "field with underscores",
			input:    "user_name:john",
			expected: []TokenType{STRING, COLON, STRING, EOF},
		},
		{
			name:     "field with hyphens",
			input:    "user-name:john",
			expected: []TokenType{STRING, COLON, STRING, EOF},
		},
		{
			name:     "multiple terms",
			input:    "quick brown fox",
			expected: []TokenType{STRING, STRING, STRING, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, expectedType := range tt.expected {
				tok := lexer.NextToken()
				if tok.Type != expectedType {
					t.Errorf("token %d: expected type %s, got %s (literal: %q)",
						i, expectedType, tok.Type, tok.Literal)
				}
			}
		})
	}
}

func TestLexer_InvalidTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{
			name:     "semicolon",
			input:    "field;value",
			expected: ILLEGAL,
		},
		{
			name:     "SQL comment dash-dash",
			input:    "field--value",
			expected: ILLEGAL,
		},
		{
			name:     "SQL comment slash-star",
			input:    "field/*value",
			expected: ILLEGAL,
		},
		{
			name:     "single ampersand",
			input:    "a & b",
			expected: ILLEGAL,
		},
		{
			name:     "single pipe",
			input:    "a | b",
			expected: ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			foundIllegal := false
			for {
				tok := lexer.NextToken()
				if tok.Type == ILLEGAL {
					foundIllegal = true
					break
				}
				if tok.Type == EOF {
					break
				}
			}
			if !foundIllegal {
				t.Errorf("expected ILLEGAL token, but didn't find one")
			}
		})
	}
}

func TestLexer_ComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, []Token)
	}{
		{
			name:  "field group query",
			input: "name:(blue OR red)",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{STRING, COLON, LPAREN, STRING, OR, STRING, RPAREN, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
				for i, tok := range tokens {
					if tok.Type != expected[i] {
						t.Errorf("token %d: expected %s, got %s", i, expected[i], tok.Type)
					}
				}
			},
		},
		{
			name:  "range with field",
			input: "rodLength:[50 TO 500]",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{STRING, COLON, LBRACKET, NUMBER, TO, NUMBER, RBRACKET, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
				for i, tok := range tokens {
					if tok.Type != expected[i] {
						t.Errorf("token %d: expected %s, got %s", i, expected[i], tok.Type)
					}
				}
			},
		},
		{
			name:  "boolean with precedence",
			input: "a AND b OR c",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{STRING, AND, STRING, OR, STRING, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
			},
		},
		{
			name:  "required and prohibited",
			input: "+required -prohibited",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{PLUS, STRING, MINUS, STRING, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
			},
		},
		{
			name:  "boost and fuzzy",
			input: "term^2~1",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{STRING, CARET, NUMBER, TILDE, NUMBER, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
			},
		},
		{
			name:  "nested groups",
			input: "(a AND (b OR c))",
			check: func(t *testing.T, tokens []Token) {
				expected := []TokenType{LPAREN, STRING, AND, LPAREN, STRING, OR, STRING, RPAREN, RPAREN, EOF}
				if len(tokens) != len(expected) {
					t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
				}
			},
		},
		{
			name:  "wildcard at start",
			input: "*widget",
			check: func(t *testing.T, tokens []Token) {
				if tokens[0].Type != WILDCARD {
					t.Errorf("expected WILDCARD, got %s", tokens[0].Type)
				}
			},
		},
		{
			name:  "wildcard at end",
			input: "widget*",
			check: func(t *testing.T, tokens []Token) {
				if tokens[0].Type != WILDCARD {
					t.Errorf("expected WILDCARD, got %s", tokens[0].Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens := lexer.AllTokens()
			tt.check(t, tokens)
		})
	}
}

func TestLexer_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escaped colon",
			input:    `test\:value`,
			expected: "test:value",
		},
		{
			name:     "escaped plus",
			input:    `test\+value`,
			expected: "test+value",
		},
		{
			name:     "escaped minus",
			input:    `test\-value`,
			expected: "test-value",
		},
		{
			name:     "escaped backslash",
			input:    `test\\value`,
			expected: `test\value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestLexer_PositionTracking(t *testing.T) {
	input := "field:value\nAND another:term"
	lexer := NewLexer(input)

	// First token should be at line 1
	tok := lexer.NextToken()
	if tok.Position.Line != 1 {
		t.Errorf("expected line 1, got %d", tok.Position.Line)
	}

	// Skip to the token after newline
	for tok.Type != AND {
		tok = lexer.NextToken()
	}

	// AND should be at line 2
	if tok.Position.Line != 2 {
		t.Errorf("expected line 2, got %d", tok.Position.Line)
	}
}

func TestLexer_QuotedStringWithEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple quoted string",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "quoted string with escaped quote",
			input:    `"hello \"world\""`,
			expected: `hello "world"`,
		},
		{
			name:     "quoted string with backslash",
			input:    `"hello\\world"`,
			expected: `hello\world`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tok := lexer.NextToken()
			if tok.Type != QUOTED_STRING {
				t.Fatalf("expected QUOTED_STRING, got %s", tok.Type)
			}
			if tok.Literal != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

func TestLexer_RegexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple regex",
			input:    "/test/",
			expected: "test",
		},
		{
			name:     "regex with character class",
			input:    "/[a-z]+/",
			expected: "[a-z]+",
		},
		{
			name:     "regex with escaped slash",
			input:    `/test\/value/`,
			expected: `test\/value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tok := lexer.NextToken()
			if tok.Type != REGEX {
				t.Fatalf("expected REGEX, got %s", tok.Type)
			}
			if tok.Literal != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

func TestLexer_CaseInsensitiveOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{
			name:     "lowercase and is string",
			input:    "and",
			expected: STRING,
		},
		{
			name:     "uppercase AND is operator",
			input:    "AND",
			expected: AND,
		},
		{
			name:     "lowercase or is string",
			input:    "or",
			expected: STRING,
		},
		{
			name:     "uppercase OR is operator",
			input:    "OR",
			expected: OR,
		},
		{
			name:     "lowercase not is string",
			input:    "not",
			expected: STRING,
		},
		{
			name:     "uppercase NOT is operator",
			input:    "NOT",
			expected: NOT,
		},
		{
			name:     "lowercase to is string",
			input:    "to",
			expected: STRING,
		},
		{
			name:     "uppercase TO is operator",
			input:    "TO",
			expected: TO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tok := lexer.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tok.Type)
			}
		})
	}
}

func TestLexer_DebugDashDash(t *testing.T) {
	input := "field--value"
	lexer := NewLexer(input)

	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		t.Logf("Type: %s, Literal: %q", tok.Type, tok.Literal)
		if tok.Type == EOF || tok.Type == ILLEGAL {
			break
		}
	}
}

func TestLexer_WildcardPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "star wildcard",
			input:    "test*",
			expected: "test*",
		},
		{
			name:     "question wildcard",
			input:    "test?",
			expected: "test?",
		},
		{
			name:     "multiple wildcards",
			input:    "t*st?",
			expected: "t*st?",
		},
		{
			name:     "only star",
			input:    "*",
			expected: "*",
		},
		{
			name:     "only question",
			input:    "?",
			expected: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tok := lexer.NextToken()
			if tok.Type != WILDCARD {
				t.Fatalf("expected WILDCARD, got %s", tok.Type)
			}
			if tok.Literal != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

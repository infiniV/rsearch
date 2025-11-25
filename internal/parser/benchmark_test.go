package parser

import (
	"testing"

	"github.com/infiniv/rsearch/internal/testdata"
)

// BenchmarkParseSimpleQuery benchmarks parsing of simple field:value queries
func BenchmarkParseSimpleQuery(b *testing.B) {
	queries := testdata.BenchmarkQueries.Simple

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseComplexQuery benchmarks parsing of complex queries with multiple operators
func BenchmarkParseComplexQuery(b *testing.B) {
	queries := testdata.BenchmarkQueries.Complex

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseLongQuery benchmarks parsing of queries with 100+ terms
func BenchmarkParseLongQuery(b *testing.B) {
	queries := testdata.BenchmarkQueries.Long

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDeepNesting benchmarks parsing of deeply nested parentheses
func BenchmarkParseDeepNesting(b *testing.B) {
	queries := testdata.BenchmarkQueries.Nested

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLexer benchmarks tokenization only (lexer performance)
func BenchmarkLexer(b *testing.B) {
	query := "productCode:13w42 AND region:ca AND status:active AND price:[100 TO 500]"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(query)
		for {
			tok := lexer.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkLexerLongQuery benchmarks lexer performance on long queries
func BenchmarkLexerLongQuery(b *testing.B) {
	query := testdata.BenchmarkQueries.Long[0]

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(query)
		for {
			tok := lexer.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkParseFieldQuery benchmarks field:value parsing
func BenchmarkParseFieldQuery(b *testing.B) {
	query := "productCode:13w42"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseRangeQuery benchmarks range query parsing
func BenchmarkParseRangeQuery(b *testing.B) {
	query := "price:[100 TO 500]"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseWildcardQuery benchmarks wildcard query parsing
func BenchmarkParseWildcardQuery(b *testing.B) {
	query := "productName:test*"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseBooleanOps benchmarks boolean operator parsing
func BenchmarkParseBooleanOps(b *testing.B) {
	query := "region:ca AND status:active OR region:ny"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseGroupQuery benchmarks grouped query parsing
func BenchmarkParseGroupQuery(b *testing.B) {
	query := "(region:ca OR region:ny) AND status:active"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseFuzzyQuery benchmarks fuzzy query parsing
func BenchmarkParseFuzzyQuery(b *testing.B) {
	query := "productCode:test~2"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseFieldGroupQuery benchmarks field:(a OR b) parsing
func BenchmarkParseFieldGroupQuery(b *testing.B) {
	query := "region:(ca OR ny OR tx)"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseExistsQuery benchmarks _exists_:field parsing
func BenchmarkParseExistsQuery(b *testing.B) {
	query := "_exists_:productCode"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(query)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

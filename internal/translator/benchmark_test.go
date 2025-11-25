package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/testdata"
)

// BenchmarkTranslateSimple benchmarks translation of simple field queries
func BenchmarkTranslateSimple(b *testing.B) {
	query := "productCode:13w42"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateComplex benchmarks translation of complex queries
func BenchmarkTranslateComplex(b *testing.B) {
	queries := testdata.BenchmarkQueries.Complex
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	// Pre-parse all queries
	asts := make([]parser.Node, len(queries))
	for i, query := range queries {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
		asts[i] = ast
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ast := asts[i%len(asts)]
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateWithManyParams benchmarks translation with 50+ parameters
func BenchmarkTranslateWithManyParams(b *testing.B) {
	// Query with many fields that will generate many parameters
	query := "region:ca OR region:ny OR region:tx OR region:fl OR region:wa " +
		"OR region:or OR region:nv OR region:az OR region:co OR region:ut " +
		"OR region:nm OR region:id OR region:mt OR region:wy OR region:nd " +
		"OR region:sd OR region:ne OR region:ks OR region:ok OR region:ar " +
		"OR region:la OR region:ms OR region:al OR region:tn OR region:ky"

	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		output, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
		if len(output.Parameters) < 20 {
			b.Fatalf("expected many parameters, got %d", len(output.Parameters))
		}
	}
}

// BenchmarkTranslateRange benchmarks range query translation
func BenchmarkTranslateRange(b *testing.B) {
	queries := []string{
		"price:[100 TO 500]",
		"price:{100 TO 500}",
		"price:[100 TO *]",
		"price:>=100",
		"price:[50 TO 500} AND quantity:[10 TO 100]",
	}

	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	// Pre-parse queries
	asts := make([]parser.Node, len(queries))
	for i, query := range queries {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
		asts[i] = ast
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ast := asts[i%len(asts)]
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateWildcard benchmarks wildcard pattern translation
func BenchmarkTranslateWildcard(b *testing.B) {
	queries := []string{
		"productName:test*",
		"productCode:*abc",
		"productName:*test*",
		"region:c?",
	}

	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	// Pre-parse queries
	asts := make([]parser.Node, len(queries))
	for i, query := range queries {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
		asts[i] = ast
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ast := asts[i%len(asts)]
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateBooleanOps benchmarks boolean operator translation
func BenchmarkTranslateBooleanOps(b *testing.B) {
	query := "region:ca AND status:active OR (region:ny AND status:pending)"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateNested benchmarks nested query translation
func BenchmarkTranslateNested(b *testing.B) {
	queries := testdata.BenchmarkQueries.Nested
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	// Pre-parse queries
	asts := make([]parser.Node, len(queries))
	for i, query := range queries {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
		asts[i] = ast
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ast := asts[i%len(asts)]
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateFieldGroup benchmarks field:(a OR b) translation
func BenchmarkTranslateFieldGroup(b *testing.B) {
	query := "region:(ca OR ny OR tx OR fl OR wa)"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateFuzzy benchmarks fuzzy search translation
func BenchmarkTranslateFuzzy(b *testing.B) {
	query := "productCode:test~2 AND region:ca"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateExists benchmarks _exists_:field translation
func BenchmarkTranslateExists(b *testing.B) {
	query := "_exists_:productCode AND region:ca"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateRegex benchmarks regex translation
func BenchmarkTranslateRegex(b *testing.B) {
	query := "productName:/test.*/ AND region:ca"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranslateBoost benchmarks boost translation (metadata only)
func BenchmarkTranslateBoost(b *testing.B) {
	query := "productCode:13w42^2.5 AND region:ca^1.5"
	schema := testdata.GetBenchmarkSchema()
	translator := NewPostgresTranslator()

	p := parser.NewParser(query)
	ast, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		output, err := translator.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
		if output.Metadata == nil {
			b.Fatal("expected boost metadata")
		}
	}
}

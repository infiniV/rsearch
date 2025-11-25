package internal

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/testdata"
	"github.com/infiniv/rsearch/internal/translator"
)

// BenchmarkFullPipeline benchmarks the complete parse -> translate flow
func BenchmarkFullPipeline(b *testing.B) {
	query := "productCode:13w42 AND region:ca AND status:active"
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Parse
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		// Translate
		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullPipelineSimple benchmarks simple queries end-to-end
func BenchmarkFullPipelineSimple(b *testing.B) {
	queries := testdata.BenchmarkQueries.Simple
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullPipelineComplex benchmarks complex queries end-to-end
func BenchmarkFullPipelineComplex(b *testing.B) {
	queries := testdata.BenchmarkQueries.Complex
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullPipelineLong benchmarks long queries end-to-end
func BenchmarkFullPipelineLong(b *testing.B) {
	query := testdata.BenchmarkQueries.Long[0]
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullPipelineNested benchmarks deeply nested queries end-to-end
func BenchmarkFullPipelineNested(b *testing.B) {
	queries := testdata.BenchmarkQueries.Nested
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentTranslations benchmarks parallel request processing
func BenchmarkConcurrentTranslations(b *testing.B) {
	query := "productCode:13w42 AND region:ca AND status:active"
	schema := testdata.GetBenchmarkSchema()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			trans := translator.NewPostgresTranslator()

			p := parser.NewParser(query)
			ast, err := p.Parse()
			if err != nil {
				b.Fatal(err)
			}

			_, err = trans.Translate(ast, schema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkConcurrentTranslationsComplex benchmarks parallel complex queries
func BenchmarkConcurrentTranslationsComplex(b *testing.B) {
	queries := testdata.BenchmarkQueries.Complex
	schema := testdata.GetBenchmarkSchema()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			query := queries[i%len(queries)]
			i++

			trans := translator.NewPostgresTranslator()

			p := parser.NewParser(query)
			ast, err := p.Parse()
			if err != nil {
				b.Fatal(err)
			}

			_, err = trans.Translate(ast, schema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPipelineWithFieldResolution benchmarks with different field naming patterns
func BenchmarkPipelineWithFieldResolution(b *testing.B) {
	queries := []string{
		"productCode:13w42",        // camelCase (exact match)
		"PRODUCTCODE:13w42",        // uppercase (case-insensitive)
		"ProductCode:13w42",        // PascalCase (case-insensitive)
		"productcode:13w42",        // lowercase (case-insensitive)
	}

	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineRangeQueries benchmarks range query processing
func BenchmarkPipelineRangeQueries(b *testing.B) {
	queries := []string{
		"price:[100 TO 500]",
		"price:{100 TO 500}",
		"price:[100 TO *]",
		"price:>=100",
		"price:<=500",
	}

	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineWildcardQueries benchmarks wildcard query processing
func BenchmarkPipelineWildcardQueries(b *testing.B) {
	queries := []string{
		"productName:test*",
		"productCode:*abc",
		"productName:*test*",
		"region:c?",
	}

	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineFieldGroups benchmarks field:(a OR b) processing
func BenchmarkPipelineFieldGroups(b *testing.B) {
	query := "region:(ca OR ny OR tx OR fl OR wa)"
	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineMixed benchmarks various query types together
func BenchmarkPipelineMixed(b *testing.B) {
	queries := []string{
		"productCode:13w42",
		"productCode:13w42 AND region:ca",
		"price:[100 TO 500]",
		"productName:test*",
		"region:(ca OR ny OR tx)",
		"status:active AND price:>=100",
		"_exists_:productCode",
		"productCode:test~2",
	}

	schema := testdata.GetBenchmarkSchema()
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineWithValidation benchmarks full pipeline with schema validation
func BenchmarkPipelineWithValidation(b *testing.B) {
	query := "productCode:13w42 AND region:ca AND status:active"
	trans := translator.NewPostgresTranslator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create fresh schema (includes validation)
		schema := testdata.GetBenchmarkSchema()

		p := parser.NewParser(query)
		ast, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}

		_, err = trans.Translate(ast, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

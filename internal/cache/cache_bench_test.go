package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/infiniv/rsearch/internal/parser"
)

// BenchmarkCacheSet benchmarks Set operations
func BenchmarkCacheSet(b *testing.B) {
	cache := NewCache(10000, time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Set(key, i)
	}
}

// BenchmarkCacheGet benchmarks Get operations
func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(10000, time.Minute)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Get(key)
	}
}

// BenchmarkCacheGetMiss benchmarks Get operations on cache misses
func BenchmarkCacheGetMiss(b *testing.B) {
	cache := NewCache(10000, time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Get(key)
	}
}

// BenchmarkCacheSetGet benchmarks mixed Set/Get operations
func BenchmarkCacheSetGet(b *testing.B) {
	cache := NewCache(10000, time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		if i%2 == 0 {
			cache.Set(key, i)
		} else {
			cache.Get(key)
		}
	}
}

// BenchmarkParseCacheSet benchmarks ParseCache Set operations
func BenchmarkParseCacheSet(b *testing.B) {
	pc := NewParseCache(10000, time.Minute)
	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("status:%d", i%1000)
		pc.Set(query, "products", node)
	}
}

// BenchmarkParseCacheGet benchmarks ParseCache Get operations
func BenchmarkParseCacheGet(b *testing.B) {
	pc := NewParseCache(10000, time.Minute)
	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		query := fmt.Sprintf("status:%d", i)
		pc.Set(query, "products", node)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("status:%d", i%1000)
		pc.Get(query, "products")
	}
}

// BenchmarkMakeKey benchmarks key generation
func BenchmarkMakeKey(b *testing.B) {
	query := "status:active AND price>100 AND category:electronics"
	schema := "products"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		MakeKey(query, schema)
	}
}

// BenchmarkCacheLRUEviction benchmarks LRU eviction performance
func BenchmarkCacheLRUEviction(b *testing.B) {
	cache := NewCache(1000, time.Minute)

	// Pre-populate to capacity
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()

	// Trigger evictions
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("newkey%d", i)
		cache.Set(key, i)
	}
}

// BenchmarkParseCacheWithComplexAST benchmarks caching complex AST nodes
func BenchmarkParseCacheWithComplexAST(b *testing.B) {
	pc := NewParseCache(10000, time.Minute)

	// Create a complex AST
	complexNode := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
			Pos:   parser.Position{},
		},
		Right: &parser.RangeQuery{
			Field:          "price",
			Start:          &parser.NumberValue{Number: "100", Pos: parser.Position{}},
			End:            &parser.TermValue{Term: "*", Pos: parser.Position{}},
			InclusiveStart: false,
			InclusiveEnd:   false,
			Pos:            parser.Position{},
		},
		Pos: parser.Position{},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("status:active AND price>%d", i%1000)
		pc.Set(query, "products", complexNode)
	}
}

// BenchmarkCacheParallel benchmarks concurrent cache operations
func BenchmarkCacheParallel(b *testing.B) {
	cache := NewCache(10000, time.Minute)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			if i%2 == 0 {
				cache.Get(key)
			} else {
				cache.Set(key, i)
			}
			i++
		}
	})
}

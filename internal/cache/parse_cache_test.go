package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/infiniv/rsearch/internal/parser"
)

// TestNewParseCache tests creation of parse cache
func TestNewParseCache(t *testing.T) {
	pc := NewParseCache(100, time.Minute)
	if pc == nil {
		t.Fatal("NewParseCache returned nil")
	}
	if pc.cache == nil {
		t.Fatal("ParseCache.cache is nil")
	}
}

// TestParseCacheSetGet tests basic set and get operations
func TestParseCacheSetGet(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	// Create a simple AST node
	node := &parser.TermQuery{
		Term: "test",
		Pos:  parser.Position{Line: 1, Column: 1},
	}

	// Set and get
	pc.Set("status:active", "products", node)
	retrieved, found := pc.Get("status:active", "products")

	if !found {
		t.Fatal("Get() returned false for existing entry")
	}

	if retrieved == nil {
		t.Fatal("Get() returned nil node")
	}

	termQuery, ok := retrieved.(*parser.TermQuery)
	if !ok {
		t.Fatal("Retrieved node is not TermQuery")
	}

	if termQuery.Term != "test" {
		t.Errorf("termQuery.Term = %s, want test", termQuery.Term)
	}
}

// TestParseCacheMiss tests cache miss scenario
func TestParseCacheMiss(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	_, found := pc.Get("nonexistent", "schema")
	if found {
		t.Error("Get() returned true for non-existent entry")
	}
}

// TestParseCacheDifferentQueries tests caching different queries
func TestParseCacheDifferentQueries(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	node1 := &parser.TermQuery{Term: "test1", Pos: parser.Position{}}
	node2 := &parser.TermQuery{Term: "test2", Pos: parser.Position{}}

	pc.Set("query1", "schema1", node1)
	pc.Set("query2", "schema1", node2)

	// Both should exist and be different
	retrieved1, found1 := pc.Get("query1", "schema1")
	retrieved2, found2 := pc.Get("query2", "schema1")

	if !found1 || !found2 {
		t.Fatal("Get() returned false for existing entries")
	}

	term1 := retrieved1.(*parser.TermQuery).Term
	term2 := retrieved2.(*parser.TermQuery).Term

	if term1 != "test1" || term2 != "test2" {
		t.Errorf("Retrieved wrong nodes: %s, %s", term1, term2)
	}
}

// TestParseCacheDifferentSchemas tests same query with different schemas
func TestParseCacheDifferentSchemas(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	node1 := &parser.TermQuery{Term: "schema1", Pos: parser.Position{}}
	node2 := &parser.TermQuery{Term: "schema2", Pos: parser.Position{}}

	pc.Set("status:active", "schema1", node1)
	pc.Set("status:active", "schema2", node2)

	// Both should exist with different values
	retrieved1, found1 := pc.Get("status:active", "schema1")
	retrieved2, found2 := pc.Get("status:active", "schema2")

	if !found1 || !found2 {
		t.Fatal("Get() returned false for existing entries")
	}

	term1 := retrieved1.(*parser.TermQuery).Term
	term2 := retrieved2.(*parser.TermQuery).Term

	if term1 != "schema1" || term2 != "schema2" {
		t.Errorf("Retrieved wrong nodes: %s, %s", term1, term2)
	}
}

// TestParseCacheComplexNodes tests caching complex AST nodes
func TestParseCacheComplexNodes(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	// Create a complex AST: (status:active AND price>100)
	node := &parser.BinaryOp{
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

	pc.Set("status:active AND price>100", "products", node)
	retrieved, found := pc.Get("status:active AND price>100", "products")

	if !found {
		t.Fatal("Get() returned false for complex node")
	}

	binaryOp, ok := retrieved.(*parser.BinaryOp)
	if !ok {
		t.Fatal("Retrieved node is not BinaryOp")
	}

	if binaryOp.Op != "AND" {
		t.Errorf("binaryOp.Op = %s, want AND", binaryOp.Op)
	}

	// Verify left side
	leftField, ok := binaryOp.Left.(*parser.FieldQuery)
	if !ok {
		t.Fatal("Left node is not FieldQuery")
	}
	if leftField.Field != "status" {
		t.Errorf("leftField.Field = %s, want status", leftField.Field)
	}

	// Verify right side
	rightRange, ok := binaryOp.Right.(*parser.RangeQuery)
	if !ok {
		t.Fatal("Right node is not RangeQuery")
	}
	if rightRange.Field != "price" {
		t.Errorf("rightRange.Field = %s, want price", rightRange.Field)
	}
}

// TestParseCacheDelete tests deletion
func TestParseCacheDelete(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}
	pc.Set("query", "schema", node)

	// Verify it exists
	_, found := pc.Get("query", "schema")
	if !found {
		t.Fatal("Entry should exist before delete")
	}

	// Delete
	pc.Delete("query", "schema")

	// Verify it's gone
	_, found = pc.Get("query", "schema")
	if found {
		t.Error("Entry should not exist after delete")
	}
}

// TestParseCacheClear tests clearing the cache
func TestParseCacheClear(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	node1 := &parser.TermQuery{Term: "test1", Pos: parser.Position{}}
	node2 := &parser.TermQuery{Term: "test2", Pos: parser.Position{}}

	pc.Set("query1", "schema1", node1)
	pc.Set("query2", "schema2", node2)

	if pc.Len() != 2 {
		t.Errorf("Len() = %d, want 2", pc.Len())
	}

	pc.Clear()

	if pc.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after Clear()", pc.Len())
	}

	_, found := pc.Get("query1", "schema1")
	if found {
		t.Error("Entry should not exist after Clear()")
	}
}

// TestParseCacheLen tests the Len method
func TestParseCacheLen(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	if pc.Len() != 0 {
		t.Errorf("Len() = %d, want 0", pc.Len())
	}

	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}

	pc.Set("query1", "schema1", node)
	if pc.Len() != 1 {
		t.Errorf("Len() = %d, want 1", pc.Len())
	}

	pc.Set("query2", "schema1", node)
	pc.Set("query3", "schema2", node)
	if pc.Len() != 3 {
		t.Errorf("Len() = %d, want 3", pc.Len())
	}

	// Setting same key should not increase length
	pc.Set("query1", "schema1", node)
	if pc.Len() != 3 {
		t.Errorf("Len() = %d, want 3 after update", pc.Len())
	}
}

// TestParseCacheTTL tests TTL expiration
func TestParseCacheTTL(t *testing.T) {
	pc := NewParseCache(10, 50*time.Millisecond)

	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}
	pc.Set("query", "schema", node)

	// Should exist immediately
	_, found := pc.Get("query", "schema")
	if !found {
		t.Fatal("Entry should exist immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = pc.Get("query", "schema")
	if found {
		t.Error("Entry should be expired")
	}
}

// TestParseCacheLRUEviction tests LRU eviction
func TestParseCacheLRUEviction(t *testing.T) {
	pc := NewParseCache(3, time.Minute)

	node := &parser.TermQuery{Term: "test", Pos: parser.Position{}}

	// Fill cache
	pc.Set("query1", "schema", node)
	pc.Set("query2", "schema", node)
	pc.Set("query3", "schema", node)

	if pc.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", pc.Len())
	}

	// Add one more - should evict query1
	pc.Set("query4", "schema", node)

	if pc.Len() != 3 {
		t.Errorf("Len() = %d, want 3 after eviction", pc.Len())
	}

	// query1 should be evicted
	_, found := pc.Get("query1", "schema")
	if found {
		t.Error("query1 should have been evicted")
	}

	// Others should exist
	if _, found := pc.Get("query2", "schema"); !found {
		t.Error("query2 should exist")
	}
}

// TestParseCacheConcurrent tests concurrent access
func TestParseCacheConcurrent(t *testing.T) {
	pc := NewParseCache(100, time.Minute)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent sets and gets
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			node := &parser.TermQuery{
				Term: "test",
				Pos:  parser.Position{},
			}

			for j := 0; j < 100; j++ {
				query := string(rune('a' + (j % 26)))
				pc.Set(query, "schema", node)
				pc.Get(query, "schema")
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	node := &parser.TermQuery{Term: "final", Pos: parser.Position{}}
	pc.Set("final", "schema", node)

	retrieved, found := pc.Get("final", "schema")
	if !found {
		t.Fatal("Cache should be functional after concurrent access")
	}

	if retrieved.(*parser.TermQuery).Term != "final" {
		t.Error("Retrieved wrong value after concurrent access")
	}
}

// TestMakeKey tests the key generation function
func TestMakeKey(t *testing.T) {
	tests := []struct {
		name        string
		query1      string
		schema1     string
		query2      string
		schema2     string
		shouldMatch bool
	}{
		{
			name:        "identical inputs",
			query1:      "status:active",
			schema1:     "products",
			query2:      "status:active",
			schema2:     "products",
			shouldMatch: true,
		},
		{
			name:        "different queries",
			query1:      "status:active",
			schema1:     "products",
			query2:      "status:inactive",
			schema2:     "products",
			shouldMatch: false,
		},
		{
			name:        "different schemas",
			query1:      "status:active",
			schema1:     "products",
			query2:      "status:active",
			schema2:     "users",
			shouldMatch: false,
		},
		{
			name:        "empty inputs",
			query1:      "",
			schema1:     "",
			query2:      "",
			schema2:     "",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := MakeKey(tt.query1, tt.schema1)
			key2 := MakeKey(tt.query2, tt.schema2)

			if tt.shouldMatch {
				if key1 != key2 {
					t.Errorf("keys should match: %s != %s", key1, key2)
				}
			} else {
				if key1 == key2 {
					t.Errorf("keys should not match: %s == %s", key1, key2)
				}
			}

			// Verify key is a valid SHA-256 hex string (64 chars)
			if len(key1) != 64 {
				t.Errorf("key length = %d, want 64", len(key1))
			}
		})
	}
}

// TestParseCacheWithRealParser tests integration with actual parser
func TestParseCacheWithRealParser(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	// Parse a real query
	p := parser.NewParser("status:active AND price>100")
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Cache it
	query := "status:active AND price>100"
	schema := "products"
	pc.Set(query, schema, node)

	// Retrieve it
	cached, found := pc.Get(query, schema)
	if !found {
		t.Fatal("Cached node not found")
	}

	// Verify it's the same structure
	binaryOp, ok := cached.(*parser.BinaryOp)
	if !ok {
		t.Fatal("Cached node is not BinaryOp")
	}

	if binaryOp.Op != "AND" {
		t.Errorf("Op = %s, want AND", binaryOp.Op)
	}

	// Verify we can use the cached node for translation
	// (We don't actually translate here, just verify structure)
	if binaryOp.Left == nil || binaryOp.Right == nil {
		t.Error("Cached node structure is incomplete")
	}
}

// TestParseCacheUpdate tests updating cached entries
func TestParseCacheUpdate(t *testing.T) {
	pc := NewParseCache(10, time.Minute)

	node1 := &parser.TermQuery{Term: "value1", Pos: parser.Position{}}
	node2 := &parser.TermQuery{Term: "value2", Pos: parser.Position{}}

	// Set initial value
	pc.Set("query", "schema", node1)

	retrieved, found := pc.Get("query", "schema")
	if !found || retrieved.(*parser.TermQuery).Term != "value1" {
		t.Fatal("Initial value incorrect")
	}

	// Update
	pc.Set("query", "schema", node2)

	retrieved, found = pc.Get("query", "schema")
	if !found || retrieved.(*parser.TermQuery).Term != "value2" {
		t.Error("Updated value incorrect")
	}

	// Length should still be 1
	if pc.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after update", pc.Len())
	}
}

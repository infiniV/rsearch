package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/infiniv/rsearch/internal/parser"
)

// ParseCache is a specialized cache for parsed AST results
type ParseCache struct {
	cache *Cache
}

// NewParseCache creates a new parse cache with the specified max size and TTL
func NewParseCache(maxSize int, ttl time.Duration) *ParseCache {
	return &ParseCache{
		cache: NewCache(maxSize, ttl),
	}
}

// Get retrieves a cached parsed AST result
func (pc *ParseCache) Get(query, schemaName string) (parser.Node, bool) {
	key := MakeKey(query, schemaName)
	value, found := pc.cache.Get(key)
	if !found {
		return nil, false
	}

	node, ok := value.(parser.Node)
	if !ok {
		// Invalid type in cache, remove it
		pc.cache.Delete(key)
		return nil, false
	}

	return node, true
}

// Set stores a parsed AST result in the cache
func (pc *ParseCache) Set(query, schemaName string, node parser.Node) {
	key := MakeKey(query, schemaName)
	pc.cache.Set(key, node)
}

// Delete removes a cached entry
func (pc *ParseCache) Delete(query, schemaName string) {
	key := MakeKey(query, schemaName)
	pc.cache.Delete(key)
}

// Clear removes all cached entries
func (pc *ParseCache) Clear() {
	pc.cache.Clear()
}

// Len returns the current number of cached entries
func (pc *ParseCache) Len() int {
	return pc.cache.Len()
}

// MakeKey creates a cache key from query string and schema name.
// Uses SHA-256 hash to create a consistent, fixed-length key.
func MakeKey(query, schemaName string) string {
	// Concatenate query and schema name with separator
	input := query + "|" + schemaName

	// Hash to create consistent key
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

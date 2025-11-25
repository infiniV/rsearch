package cache

import (
	"sync"
	"testing"
	"time"
)

// TestNewCache verifies cache initialization
func TestNewCache(t *testing.T) {
	tests := []struct {
		name    string
		maxSize int
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "valid cache",
			maxSize: 100,
			ttl:     time.Minute,
			wantErr: false,
		},
		{
			name:    "zero max size",
			maxSize: 0,
			ttl:     time.Minute,
			wantErr: true,
		},
		{
			name:    "negative max size",
			maxSize: -1,
			ttl:     time.Minute,
			wantErr: true,
		},
		{
			name:    "zero TTL (no expiration)",
			maxSize: 100,
			ttl:     0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(tt.maxSize, tt.ttl)
			if tt.wantErr {
				if cache != nil {
					t.Errorf("NewCache() should return nil for invalid params")
				}
			} else {
				if cache == nil {
					t.Errorf("NewCache() returned nil for valid params")
				}
				if cache.maxSize != tt.maxSize {
					t.Errorf("maxSize = %d, want %d", cache.maxSize, tt.maxSize)
				}
				if cache.ttl != tt.ttl {
					t.Errorf("ttl = %v, want %v", cache.ttl, tt.ttl)
				}
			}
		})
	}
}

// TestCacheSetGet tests basic set and get operations
func TestCacheSetGet(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Test setting and getting a value
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	if !found {
		t.Error("Get() returned false for existing key")
	}
	if value != "value1" {
		t.Errorf("Get() = %v, want %v", value, "value1")
	}

	// Test getting non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Get() returned true for non-existent key")
	}

	// Test setting different types
	cache.Set("int", 42)
	value, found = cache.Get("int")
	if !found || value != 42 {
		t.Errorf("Get(int) = %v, %v, want 42, true", value, found)
	}

	cache.Set("struct", struct{ Name string }{"test"})
	value, found = cache.Get("struct")
	if !found {
		t.Error("Get(struct) returned false")
	}
}

// TestCacheLen tests the Len method
func TestCacheLen(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	if cache.Len() != 0 {
		t.Errorf("Len() = %d, want 0", cache.Len())
	}

	cache.Set("key1", "value1")
	if cache.Len() != 1 {
		t.Errorf("Len() = %d, want 1", cache.Len())
	}

	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	if cache.Len() != 3 {
		t.Errorf("Len() = %d, want 3", cache.Len())
	}

	// Setting same key should not increase length
	cache.Set("key1", "new_value")
	if cache.Len() != 3 {
		t.Errorf("Len() = %d, want 3 after update", cache.Len())
	}
}

// TestCacheDelete tests the Delete method
func TestCacheDelete(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Delete("key1")
	_, found := cache.Get("key1")
	if found {
		t.Error("Get() returned true after Delete()")
	}

	if cache.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after delete", cache.Len())
	}

	// Deleting non-existent key should be safe
	cache.Delete("nonexistent")
	if cache.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after deleting non-existent", cache.Len())
	}
}

// TestCacheClear tests the Clear method
func TestCacheClear(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after Clear()", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Get() returned true after Clear()")
	}
}

// TestCacheLRUEviction tests LRU eviction when cache is full
func TestCacheLRUEviction(t *testing.T) {
	cache := NewCache(3, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", cache.Len())
	}

	// Add one more - should evict key1 (oldest)
	cache.Set("key4", "value4")

	if cache.Len() != 3 {
		t.Errorf("Len() = %d, want 3 after eviction", cache.Len())
	}

	// key1 should be evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("key1 should have been evicted")
	}

	// Other keys should still exist
	if _, found := cache.Get("key2"); !found {
		t.Error("key2 should exist")
	}
	if _, found := cache.Get("key3"); !found {
		t.Error("key3 should exist")
	}
	if _, found := cache.Get("key4"); !found {
		t.Error("key4 should exist")
	}
}

// TestCacheLRUAccessOrder tests that accessing an item moves it to front
func TestCacheLRUAccessOrder(t *testing.T) {
	cache := NewCache(3, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Access key1 to move it to front
	cache.Get("key1")

	// Add key4 - should evict key2 (oldest, not key1)
	cache.Set("key4", "value4")

	// key2 should be evicted
	_, found := cache.Get("key2")
	if found {
		t.Error("key2 should have been evicted")
	}

	// key1 should still exist because we accessed it
	if _, found := cache.Get("key1"); !found {
		t.Error("key1 should exist after being accessed")
	}
}

// TestCacheTTLExpiration tests TTL-based expiration
func TestCacheTTLExpiration(t *testing.T) {
	cache := NewCache(10, 50*time.Millisecond)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")

	// Should exist immediately
	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Error("key1 should exist immediately after Set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not exist after TTL
	_, found = cache.Get("key1")
	if found {
		t.Error("key1 should have expired after TTL")
	}
}

// TestCacheNoTTL tests cache with no TTL (zero TTL means no expiration)
func TestCacheNoTTL(t *testing.T) {
	cache := NewCache(10, 0)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Should still exist (no TTL)
	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Error("key1 should exist (no TTL)")
	}
}

// TestCacheConcurrentAccess tests concurrent operations
func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache(100, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id*numOperations+j)%26))
				cache.Set(key, id*numOperations+j)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id*numOperations+j)%26))
				cache.Get(key)
			}
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < numGoroutines/5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/5; j++ {
				key := string(rune('a' + (id*numOperations/5+j)%26))
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// No race conditions should occur (test should complete)
	t.Log("Concurrent access test completed successfully")
}

// TestCacheUpdateValue tests updating existing values
func TestCacheUpdateValue(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	value, found := cache.Get("key1")
	if !found {
		t.Error("key1 should exist")
	}
	if value != "value2" {
		t.Errorf("Get() = %v, want value2", value)
	}

	// Length should still be 1
	if cache.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after update", cache.Len())
	}
}

// TestCacheLRUWithUpdates tests that updating a value moves it to front
func TestCacheLRUWithUpdates(t *testing.T) {
	cache := NewCache(3, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Update key1 (should move to front)
	cache.Set("key1", "new_value1")

	// Add key4 - should evict key2 (oldest)
	cache.Set("key4", "value4")

	if _, found := cache.Get("key2"); found {
		t.Error("key2 should have been evicted")
	}

	if value, found := cache.Get("key1"); !found || value != "new_value1" {
		t.Error("key1 should exist with new value")
	}
}

// TestCacheEvictionMultiple tests multiple evictions
func TestCacheEvictionMultiple(t *testing.T) {
	cache := NewCache(5, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Fill cache
	for i := 0; i < 5; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	// Add 5 more items - should evict first 5
	for i := 5; i < 10; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	// First 5 should be evicted
	for i := 0; i < 5; i++ {
		if _, found := cache.Get(string(rune('a' + i))); found {
			t.Errorf("key %c should have been evicted", 'a'+i)
		}
	}

	// Last 5 should exist
	for i := 5; i < 10; i++ {
		if _, found := cache.Get(string(rune('a' + i))); !found {
			t.Errorf("key %c should exist", 'a'+i)
		}
	}

	if cache.Len() != 5 {
		t.Errorf("Len() = %d, want 5", cache.Len())
	}
}

// TestCacheMixedOperations tests a mix of operations
func TestCacheMixedOperations(t *testing.T) {
	cache := NewCache(10, time.Minute)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Add some items
	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Get and verify
	if val, found := cache.Get("key1"); !found || val != 1 {
		t.Error("key1 should be 1")
	}

	// Update
	cache.Set("key2", 22)

	// Delete
	cache.Delete("key3")

	// Verify state
	if cache.Len() != 2 {
		t.Errorf("Len() = %d, want 2", cache.Len())
	}

	if val, found := cache.Get("key2"); !found || val != 22 {
		t.Error("key2 should be 22")
	}

	if _, found := cache.Get("key3"); found {
		t.Error("key3 should be deleted")
	}
}

// TestCacheStressTest is a more intensive stress test
func TestCacheStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	cache := NewCache(1000, time.Second)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (j % 26)))
				switch j % 4 {
				case 0:
					cache.Set(key, j)
				case 1:
					cache.Get(key)
				case 2:
					cache.Delete(key)
				case 3:
					_ = cache.Len()
				}
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	cache.Set("test", "value")
	if val, found := cache.Get("test"); !found || val != "value" {
		t.Error("cache should still be functional after stress test")
	}
}

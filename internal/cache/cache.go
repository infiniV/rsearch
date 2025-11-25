package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache is a thread-safe LRU cache with TTL support
type Cache struct {
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
	items   map[string]*list.Element
	lruList *list.List
}

// entry represents a cache entry with value and expiration
type entry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewCache creates a new LRU cache with the specified max size and TTL.
// maxSize must be greater than 0. TTL of 0 means no expiration.
// Returns nil if maxSize is invalid.
func NewCache(maxSize int, ttl time.Duration) *Cache {
	if maxSize <= 0 {
		return nil
	}

	return &Cache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*list.Element),
		lruList: list.New(),
	}
}

// Get retrieves a value from the cache.
// Returns (value, true) if found and not expired, (nil, false) otherwise.
// Accessing an item moves it to the front of the LRU list.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		return nil, false
	}

	entry := element.Value.(*entry)

	// Check if expired
	if c.ttl > 0 && time.Now().After(entry.expiresAt) {
		// Remove expired entry
		c.removeElement(element)
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(element)

	return entry.value, true
}

// Set adds or updates a value in the cache.
// If the cache is at capacity and the key doesn't exist, the least recently used item is evicted.
// If the key exists, its value is updated and it's moved to the front.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate expiration time
	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = time.Now().Add(c.ttl)
	}

	// Check if key already exists
	if element, exists := c.items[key]; exists {
		// Update existing entry
		entry := element.Value.(*entry)
		entry.value = value
		entry.expiresAt = expiresAt
		c.lruList.MoveToFront(element)
		return
	}

	// Add new entry
	entry := &entry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}

	element := c.lruList.PushFront(entry)
	c.items[key] = element

	// Evict oldest if at capacity
	if c.lruList.Len() > c.maxSize {
		c.evictOldest()
	}
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.removeElement(element)
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList.Init()
}

// Len returns the current number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// evictOldest removes the least recently used item from the cache.
// Must be called with lock held.
func (c *Cache) evictOldest() {
	element := c.lruList.Back()
	if element != nil {
		c.removeElement(element)
	}
}

// removeElement removes an element from the cache.
// Must be called with lock held.
func (c *Cache) removeElement(element *list.Element) {
	c.lruList.Remove(element)
	entry := element.Value.(*entry)
	delete(c.items, entry.key)
}

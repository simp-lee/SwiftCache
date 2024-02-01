package swiftcache

import (
	"container/list"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"log"
	"sync"
	"time"
)

// CacheConfig is used to configure a cache instance.
type CacheConfig struct {
	SegmentCount      int                // Number of segments to reduce lock contention
	MaxCacheSize      int                // Maximum size for each cache segment
	DefaultExpiration time.Duration      // Default expiration time for cache items
	HashFunc          func() hash.Hash32 // Hash function to distribute keys across segments.
	EvictionPolicy    string             // Eviction policy: "LRU" or "FIFO".
}

const (
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
	// For use with functions that take an expiration time.
	NoExpiration          time.Duration = -1
	DefaultSegmentCount                 = 512   // Default number of segments to reduce lock contention
	MaxCacheSize                        = 1000  // Default maximum size for each cache segment
	DefaultEvictionPolicy               = "LRU" // Default eviction policy: "LRU".
)

// Item defines an item in the cache
type Item struct {
	Value      interface{}   // Value of the cache item
	Expiration int64         // Expiration time in nanoseconds
	node       *list.Element // Used for LRU to point to the node in the list.
}

// Expired checks if the cache item is expired
func (item *Item) Expired() bool {
	return item.Expiration != 0 && time.Now().UnixNano() > item.Expiration
}

// Segment represents a segment of the cache
type Segment struct {
	items   map[string]*Item // Map to store cache items
	queue   *list.List       // Used for both FIFO and LRU. The usage depends on the eviction policy.
	lock    sync.RWMutex     // Read/Write lock for concurrent access
	size    int              // Current size of the cache segment
	maxSize int              // Max size of the cache segment
	cache   *Cache           // Reference to the parent Cache.
}

// newSegment creates a new cache segment
func newSegment(maxSize int, cache *Cache) *Segment {
	return &Segment{
		items:   make(map[string]*Item),
		queue:   list.New(),
		size:    0,
		maxSize: maxSize,
		cache:   cache,
	}
}

// Cache is a structure holding multiple segments
type Cache struct {
	segments          []*Segment                // Slice of cache segments
	segmentCount      int                       // Number of segments
	maxCacheSize      int                       // Maximum size per segment
	defaultExpiration time.Duration             // Default expiration time for segment items
	hashFunc          func() hash.Hash32        // Hash function to distribute keys across segments.
	onEvicted         func(string, interface{}) // Optional callback for evicted items.
	evictionPolicy    string                    // Store the eviction policy here.
	lock              sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache(options ...CacheConfig) (*Cache, error) {
	config := CacheConfig{
		SegmentCount:      DefaultSegmentCount, // Number of segments to reduce lock contention
		MaxCacheSize:      MaxCacheSize,        // Maximum size for each cache segment
		DefaultExpiration: DefaultExpiration,
		HashFunc:          fnv.New32,
		EvictionPolicy:    DefaultEvictionPolicy,
	}

	if len(options) > 0 {
		userConfig := options[0]
		if userConfig.SegmentCount > 0 {
			config.SegmentCount = userConfig.SegmentCount
		}
		if userConfig.MaxCacheSize > 0 {
			config.MaxCacheSize = userConfig.MaxCacheSize
		}
		if userConfig.DefaultExpiration >= 0 {
			config.DefaultExpiration = userConfig.DefaultExpiration
		}
		if userConfig.HashFunc != nil {
			config.HashFunc = userConfig.HashFunc
		}
		if userConfig.EvictionPolicy != "" {
			config.EvictionPolicy = userConfig.EvictionPolicy
		}
	}

	// Validate and set defaults for config
	if config.SegmentCount <= 0 {
		config.SegmentCount = DefaultSegmentCount
	}
	if config.MaxCacheSize <= 0 {
		config.MaxCacheSize = MaxCacheSize
	}
	if config.HashFunc == nil {
		config.HashFunc = fnv.New32
	}

	if config.DefaultExpiration < -1 {
		config.DefaultExpiration = DefaultExpiration
	}

	if config.SegmentCount&(config.SegmentCount-1) != 0 {
		return nil, fmt.Errorf("cache segment count must be a power of 2")
	}

	c := &Cache{
		segments:          make([]*Segment, config.SegmentCount),
		segmentCount:      config.SegmentCount,
		maxCacheSize:      config.MaxCacheSize,
		defaultExpiration: config.DefaultExpiration,
		hashFunc:          config.HashFunc,
		evictionPolicy:    config.EvictionPolicy,
	}
	for i := range c.segments {
		c.segments[i] = newSegment(c.maxCacheSize, c)
	}

	return c, nil
}

// set sets a key-value pair in the cache
func (s *Segment) set(key string, value interface{}, ttl, defaultExpiration time.Duration) {
	var expiration int64

	if ttl == 0 {
		expiration = 0
	} else if ttl == DefaultExpiration {
		expiration = time.Now().Add(defaultExpiration).UnixNano()
	} else if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if itm, ok := s.items[key]; ok {
		// Update existing item
		itm.Value = value
		itm.Expiration = expiration

		s.queue.MoveToFront(itm.node) // Move to front as it's recently updated

		return
	}

	// Create a new item
	itm := &Item{
		Value:      value,
		Expiration: expiration,
	}

	itm.node = s.queue.PushFront(key) // Store key in LRU/FIFO list

	s.items[key] = itm
	s.size++

	// Ensure cache size does not exceed max limit
	for s.size > s.maxSize {
		s.removeOldest()
	}
}

// get retrieves a value for a key from the cache. It also updates the LRU list
func (s *Segment) get(key string) (interface{}, bool) {
	if s.cache.evictionPolicy == "LRU" {
		s.lock.Lock()
		defer s.lock.Unlock()

		item, exists := s.items[key]

		if !exists {
			return nil, false
		}

		// If the item exists but is expired, remove it
		if item.Expired() {
			s.removeKey(key)
			return nil, false
		}
		// If the item exists and is not expired, move it to the front of LRU list
		s.queue.MoveToFront(item.node)

		return item.Value, true
	} else if s.cache.evictionPolicy == "FIFO" {
		s.lock.RLock()
		item, exists := s.items[key]
		s.lock.RUnlock()

		if !exists {
			return nil, false
		}

		// If the item exists but is expired, remove it
		if item.Expired() {
			s.lock.Lock()
			s.removeKey(key)
			s.lock.Unlock()
			return nil, false
		}

		return item.Value, true
	}

	return nil, false
}

// removeKey removes a key from the cache
func (s *Segment) removeKey(key string) {
	if item, exists := s.items[key]; exists {
		if s.cache.onEvicted != nil {
			s.cache.onEvicted(key, item.Value)
		}

		s.queue.Remove(item.node) // Remove item.node from LRU/FIFO

		delete(s.items, key) // Remove item from map
		s.size--             // Update the segment size
	}
}

// Delete removes a key from the cache
func (s *Segment) delete(key string) {
	s.lock.Lock()
	s.removeKey(key)
	s.lock.Unlock()
}

// removeOldest removes the least recently used item from the cache
func (s *Segment) removeOldest() {
	if oldest := s.queue.Back(); oldest != nil {
		s.removeKey(oldest.Value.(string))
	}
}

// getWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (s *Segment) getWithExpiration(key string) (interface{}, time.Time, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	item, exists := s.items[key]
	if !exists || item.Expired() {
		return nil, time.Time{}, false
	}
	expiration := time.Time{}
	if item.Expiration > 0 {
		expiration = time.Unix(0, item.Expiration)
	}
	return item.Value, expiration, true
}

func (s *Segment) itemCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items)
}

func (s *Segment) getItems() map[string]interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make(map[string]interface{})
	for key, item := range s.items {
		if !item.Expired() {
			result[key] = item.Value
		}
	}
	return result
}

// increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (s *Segment) increment(k string, n int64) error {
	s.lock.Lock()         // 使用正确的锁名称
	defer s.lock.Unlock() // 使用 defer 确保锁一定会被释放

	v, found := s.items[k]
	if !found || v.Expired() {
		return fmt.Errorf("item %s not found or expired", k)
	}

	switch val := v.Value.(type) {
	case int:
		v.Value = val + int(n)
	case int8:
		v.Value = val + int8(n)
	case int16:
		v.Value = val + int16(n)
	case int32:
		v.Value = val + int32(n)
	case int64:
		v.Value = val + n
	case uint:
		v.Value = val + uint(n)
	case uintptr:
		v.Value = val + uintptr(n)
	case uint8:
		v.Value = val + uint8(n)
	case uint16:
		v.Value = val + uint16(n)
	case uint32:
		v.Value = val + uint32(n)
	case uint64:
		v.Value = val + uint64(n)
	case float32:
		v.Value = val + float32(n)
	case float64:
		v.Value = val + float64(n)
	default:
		return fmt.Errorf("the value for %s is not a number or not suitable for increment", k)
	}

	s.items[k] = v
	return nil
}

// decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (s *Segment) decrement(k string, n int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, found := s.items[k]
	if !found || v.Expired() {
		return fmt.Errorf("item %s not found or expired", k)
	}
	switch val := v.Value.(type) {
	case int:
		v.Value = val - int(n)
	case int8:
		v.Value = val - int8(n)
	case int16:
		v.Value = val - int16(n)
	case int32:
		v.Value = val - int32(n)
	case int64:
		v.Value = val - n
	case uint:
		if uint(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uint(n)
	case uintptr:
		if uintptr(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uintptr(n)
	case uint8:
		if uint8(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uint8(n)
	case uint16:
		if uint16(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uint16(n)
	case uint32:
		if uint32(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uint32(n)
	case uint64:
		if uint64(n) > val {
			return fmt.Errorf("decrement would result in negative value for key %s", k)
		}
		v.Value = val - uint64(n)
	case float32:
		v.Value = val - float32(n)
	case float64:
		v.Value = val - float64(n)
	default:
		return fmt.Errorf("the value for %s is not a number or not suitable for decrement", k)
	}
	s.items[k] = v
	return nil
}

// clear removes all items from the segment.
func (s *Segment) clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = make(map[string]*Item)
	s.queue.Init()
	s.size = 0
}

// getSegment computes the segment for a given key.
// It uses bit manipulation (bitwise AND operation) instead of modulo operation for efficiency.
// Bitwise operations are generally faster than arithmetic operations like modulo,
// especially when dealing with large amounts of data.
func (c *Cache) getSegment(key string) *Segment {
	hasher := c.hashFunc()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		log.Printf("Error hashing key: %v", err)
		return nil
	}

	// Using bitwise AND operation for better performance.
	// This requires that segmentCount is a power of 2.
	// c.segments[hasher.Sum32()%uint32(c.segmentCount)]
	return c.segments[hasher.Sum32()&(uint32(c.segmentCount)-1)]
}

// Set sets a key-value pair in the cache (public interface)
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	segment := c.getSegment(key)
	if segment != nil {
		segment.set(key, value, ttl, c.defaultExpiration)
	}
}

// Get retrieves a value for a key from the cache (public interface)
func (c *Cache) Get(key string) (interface{}, bool) {
	segment := c.getSegment(key)
	return segment.get(key)
}

// Delete removes a key from the cache (public interface)
func (c *Cache) Delete(key string) {
	segment := c.getSegment(key)
	if segment != nil {
		segment.delete(key)
	}
}

// GetWithExpiration returns an item and its expiration time from the cache.
func (c *Cache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
	segment := c.getSegment(key)
	if segment == nil {
		return nil, time.Time{}, false
	}
	return segment.getWithExpiration(key)
}

// ItemCount returns the number of items in the cache.
func (c *Cache) ItemCount() int {
	count := 0
	for _, segment := range c.segments {
		count += segment.itemCount()
	}
	return count
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (c *Cache) Items() map[string]interface{} {
	items := make(map[string]interface{})
	for _, segment := range c.segments {
		segmentItems := segment.getItems()
		for k, v := range segmentItems {
			items[k] = v
		}
	}
	return items
}

// Item retrieves an item from the cache, along with its existence.
// It returns a pointer to the Item and a boolean indicating whether the item was found.
func (c *Cache) Item(key string) (*Item, bool) {
	segment := c.getSegment(key)
	segment.lock.RLock()
	defer segment.lock.RUnlock()

	item, found := segment.items[key]
	return item, found
}

// Increment increases the value of an item by n.
func (c *Cache) Increment(k string, n int64) error {
	segment := c.getSegment(k)
	if segment == nil {
		return errors.New("key not found")
	}
	return segment.increment(k, n)
}

// Decrement decreases the value of an item by n.
func (c *Cache) Decrement(k string, n int64) error {
	segment := c.getSegment(k)
	if segment == nil {
		return errors.New("key not found")
	}
	return segment.decrement(k, n)
}

// Flush clears all cached items from the cache.
func (c *Cache) Flush() {
	for _, segment := range c.segments {
		segment.clear()
	}
}

// OnEvicted sets an (optional) function that is called with the key and value
// when an item is evicted from the cache. (Including when it is deleted manually,
// but not when it is overwritten.) Set to nil to disable.
func (c *Cache) OnEvicted(f func(string, interface{})) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.onEvicted = f
}

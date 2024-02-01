package swiftcache

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	tc, _ := NewCache()

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, 1*time.Minute)
	tc.Set("b", "b", 0)
	tc.Set("c", 3.5, 0)

	x, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	var found bool

	tc, _ := NewCache()
	tc.Set("a", 1, 50*time.Millisecond)
	tc.Set("b", 2, 0)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 80*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(40 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestIncrementWithInt(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tint", 1, DefaultExpiration)
	err := tc.Increment("tint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint")
	if !found {
		t.Error("tint was not found")
	}
	if x.(int) != 3 {
		t.Error("tint is not 3:", x)
	}
}

func TestIncrementWithInt8(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tint8", int8(1), DefaultExpiration)
	err := tc.Increment("tint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint8")
	if !found {
		t.Error("tint8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("tint8 is not 3:", x)
	}
}

func TestIncrementWithInt16(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tint16", int16(1), DefaultExpiration)
	err := tc.Increment("tint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint16")
	if !found {
		t.Error("tint16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("tint16 is not 3:", x)
	}
}

func TestIncrementWithInt32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tint32", int32(1), DefaultExpiration)
	err := tc.Increment("tint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint32")
	if !found {
		t.Error("tint32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("tint32 is not 3:", x)
	}
}

func TestIncrementWithInt64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tint64", int64(1), DefaultExpiration)
	err := tc.Increment("tint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint64")
	if !found {
		t.Error("tint64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("tint64 is not 3:", x)
	}
}

func TestIncrementWithUint(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuint", uint(1), DefaultExpiration)
	err := tc.Increment("tuint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint")
	if !found {
		t.Error("tuint was not found")
	}
	if x.(uint) != 3 {
		t.Error("tuint is not 3:", x)
	}
}

func TestIncrementWithUintptr(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuintptr", uintptr(1), DefaultExpiration)
	err := tc.Increment("tuintptr", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuintptr")
	if !found {
		t.Error("tuintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("tuintptr is not 3:", x)
	}
}

func TestIncrementWithUint8(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuint8", uint8(1), DefaultExpiration)
	err := tc.Increment("tuint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint8")
	if !found {
		t.Error("tuint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("tuint8 is not 3:", x)
	}
}

func TestIncrementWithUint16(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuint16", uint16(1), DefaultExpiration)
	err := tc.Increment("tuint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuint16")
	if !found {
		t.Error("tuint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("tuint16 is not 3:", x)
	}
}

func TestIncrementWithUint32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuint32", uint32(1), DefaultExpiration)
	err := tc.Increment("tuint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint32")
	if !found {
		t.Error("tuint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("tuint32 is not 3:", x)
	}
}

func TestIncrementWithUint64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("tuint64", uint64(1), DefaultExpiration)
	err := tc.Increment("tuint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuint64")
	if !found {
		t.Error("tuint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("tuint64 is not 3:", x)
	}
}

func TestIncrementWithFloat32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("float32", float32(1.5), DefaultExpiration)
	err := tc.Increment("float32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestIncrementWithFloat64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("float64", float64(1.5), DefaultExpiration)
	err := tc.Increment("float64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestDecrementInt8(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("int8", int8(5), DefaultExpiration)
	err := tc.Decrement("int8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int8")
	if !found {
		t.Error("int8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("int8 is not 3:", x)
	}
}

func TestDecrementInt16(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("int16", int16(5), DefaultExpiration)
	err := tc.Decrement("int16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int16")
	if !found {
		t.Error("int16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("int16 is not 3:", x)
	}
}

func TestDecrementInt32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("int32", int32(5), DefaultExpiration)
	err := tc.Decrement("int32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int32")
	if !found {
		t.Error("int32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("int32 is not 3:", x)
	}
}

func TestDecrementInt64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("int64", int64(5), DefaultExpiration)
	err := tc.Decrement("int64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int64")
	if !found {
		t.Error("int64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("int64 is not 3:", x)
	}
}

func TestDecrementUint(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint", uint(5), DefaultExpiration)
	err := tc.Decrement("uint", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint")
	if !found {
		t.Error("uint was not found")
	}
	if x.(uint) != 3 {
		t.Error("uint is not 3:", x)
	}
}

func TestDecrementUintptr(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uintptr", uintptr(5), DefaultExpiration)
	err := tc.Decrement("uintptr", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uintptr")
	if !found {
		t.Error("uintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("uintptr is not 3:", x)
	}
}

func TestDecrementUint8(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint8", uint8(5), DefaultExpiration)
	err := tc.Decrement("uint8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint8")
	if !found {
		t.Error("uint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("uint8 is not 3:", x)
	}
}

func TestDecrementUint16(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint16", uint16(5), DefaultExpiration)
	err := tc.Decrement("uint16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint16")
	if !found {
		t.Error("uint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("uint16 is not 3:", x)
	}
}

func TestDecrementUint32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint32", uint32(5), DefaultExpiration)
	err := tc.Decrement("uint32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint32")
	if !found {
		t.Error("uint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("uint32 is not 3:", x)
	}
}

func TestDecrementUint64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint64", uint64(5), DefaultExpiration)
	err := tc.Decrement("uint64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint64")
	if !found {
		t.Error("uint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("uint64 is not 3:", x)
	}
}

func TestDecrementFloat32(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("float32", float32(5), DefaultExpiration)
	err := tc.Decrement("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3 {
		t.Error("float32 is not 3:", x)
	}
}

func TestDecrementFloat64(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("float64", float64(5), DefaultExpiration)
	err := tc.Decrement("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3 {
		t.Error("float64 is not 3:", x)
	}
}

func TestDelete(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Delete("foo")
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestItemCount(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}
}

func TestFlush(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
	x, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestIncrementOverflowInt(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("int8", int8(127), DefaultExpiration)
	err := tc.Increment("int8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("int8")
	_int8 := x.(int8)
	if _int8 != -128 {
		t.Error("int8 did not overflow as expected; value:", _int8)
	}

}

func TestIncrementOverflowUint(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint8", uint8(255), DefaultExpiration)
	err := tc.Increment("uint8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("uint8")
	_uint8 := x.(uint8)
	if _uint8 != 0 {
		t.Error("uint8 did not overflow as expected; value:", _uint8)
	}
}

func TestDecrementUnderflowUint(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("uint8", uint8(0), DefaultExpiration)
	err := tc.Decrement("uint8", 1)
	if err == nil {
		t.Error("Error decrementing int8: decrement would result in negative value for key uint8")
	}
}

func TestOnEvicted(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("foo", 3, DefaultExpiration)
	if tc.onEvicted != nil {
		t.Fatal("tc.onEvicted is not nil")
	}
	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
		tc.Set("bar", 4, DefaultExpiration)
	})
	tc.Delete("foo")
	x, _ := tc.Get("bar")
	if !works {
		t.Error("works bool not true")
	}
	if x.(int) != 4 {
		t.Error("bar was not 4")
	}
}

func TestGetWithExpiration(t *testing.T) {
	tc, err := NewCache()
	if err != nil {
		t.Error("setting failed: ", err)
	}

	a, expiration, found := tc.GetWithExpiration("a")
	if found || a != nil || !expiration.IsZero() {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, expiration, found := tc.GetWithExpiration("b")
	if found || b != nil || !expiration.IsZero() {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, expiration, found := tc.GetWithExpiration("c")
	if found || c != nil || !expiration.IsZero() {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)
	tc.Set("d", 1, NoExpiration)
	tc.Set("e", 1, 50*time.Millisecond)

	x, expiration, found := tc.GetWithExpiration("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for a is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for b is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for c is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("d")
	if !found {
		t.Error("d was not found while getting d2")
	}
	if x == nil {
		t.Error("x for d is nil")
	} else if d2 := x.(int); d2+2 != 3 {
		t.Error("d (which should be 1) plus 2 does not equal 3; value:", d2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for d is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("e")
	if !found {
		t.Error("e was not found while getting e2")
	}
	if x == nil {
		t.Error("x for e is nil")
	} else if e2 := x.(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}

	item, found := tc.Item("e")

	if expiration.UnixNano() != item.Expiration {
		t.Error("expiration for e is not the correct time")
	}
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestStorePointerToStruct(t *testing.T) {
	tc, _ := NewCache()
	tc.Set("foo", &TestStruct{Num: 1}, 0)
	x, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := x.(*TestStruct)
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

// Test LRU eviction policy
func TestCacheLRUEviction(t *testing.T) {
	maxSize := 5
	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   1,
		MaxCacheSize:   maxSize,
		EvictionPolicy: "LRU",
	})

	// Populate cache to its maximum capacity
	for i := 0; i < maxSize; i++ {
		_cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), NoExpiration)
	}

	// Re-access the first key to make it recently used
	_cache.Get("key0")

	// Add a new element to trigger eviction
	_cache.Set("key_new", "value_new", NoExpiration)

	// Check if the earliest key ("key1") was evicted
	if _, found := _cache.Get("key1"); found {
		t.Errorf("LRU eviction failed: key1 should have been evicted")
	}
}

// Test FIFO eviction policy
func TestCacheFIFOEviction(t *testing.T) {
	maxSize := 5
	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   1,
		MaxCacheSize:   maxSize,
		EvictionPolicy: "FIFO",
	})

	// Populate cache to its maximum capacity
	for i := 0; i < maxSize; i++ {
		_cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), NoExpiration)
	}

	// Add a new element to trigger eviction
	_cache.Set("key_new", "value_new", NoExpiration)

	// Check if the earliest key ("key0") was evicted
	if _, found := _cache.Get("key0"); found {
		t.Errorf("FIFO eviction failed: key0 should have been evicted")
	}
}

func TestCacheLRUAndFIFOEviction(t *testing.T) {
	segmentCount := 4
	maxSize := 5
	testEvictionPolicy(t, segmentCount, maxSize, "LRU")
	testEvictionPolicy(t, segmentCount, maxSize, "FIFO")
}

// Returns the segment index for a given key
func getSegmentIndex(c *Cache, key string) int {
	hasher := c.hashFunc()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		log.Printf("Error hashing key: %v", err)
		return -1
	}
	return int(hasher.Sum32() & uint32(c.segmentCount-1))
}

// Check if all segments are filled
func allSegmentsFilled(segmentFill []int, maxSize int) bool {
	for _, fill := range segmentFill {
		if fill < maxSize {
			return false
		}
	}
	return true
}

// printCacheContents prints the contents of the cache for debugging
func printCacheContents(c *Cache) {
	fmt.Println("Cache contents:")
	for i, segment := range c.segments {
		fmt.Printf("Segment %d:\n", i)
		segment.lock.RLock()
		for key, item := range segment.items {
			fmt.Printf("  Key: %s, Value: %v\n", key, item.Value)
		}
		segment.lock.RUnlock()
	}
	fmt.Println()
}

func testEvictionPolicy(t *testing.T, segmentCount, maxSize int, policy string) {
	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   segmentCount,
		MaxCacheSize:   maxSize,
		EvictionPolicy: policy,
	})

	segmentFill := make([]int, segmentCount)
	firstTwoKeysInSegment := make(map[int][]string)
	var i int
	for {
		key := fmt.Sprintf("key%d", i)
		segmentIndex := getSegmentIndex(_cache, key)
		if segmentIndex != -1 && segmentFill[segmentIndex] < maxSize {
			if len(firstTwoKeysInSegment[segmentIndex]) < 2 {
				firstTwoKeysInSegment[segmentIndex] = append(firstTwoKeysInSegment[segmentIndex], key)
			}
			_cache.Set(key, "value", NoExpiration)
			segmentFill[segmentIndex]++
			if allSegmentsFilled(segmentFill, maxSize) {
				break
			}
		}
		i++
	}

	// Add a extraKey to trigger eviction
	extraKey := fmt.Sprintf("key%d", i+1)
	segmentIndexOfExtraKey := getSegmentIndex(_cache, extraKey)
	firstKey := firstTwoKeysInSegment[segmentIndexOfExtraKey][0]

	if policy == "LRU" {
		// Access the first key to make it recently used
		_cache.Get(firstKey)

		// Add a new element to trigger eviction
		_cache.Set(extraKey, "extra_value", NoExpiration)

		// Check if the second key was evicted
		secondKey := firstTwoKeysInSegment[segmentIndexOfExtraKey][1]
		_, found := _cache.Get(secondKey)
		if found {
			t.Errorf("LRU eviction failed: %s should have been evicted in segment %d", secondKey, segmentIndexOfExtraKey)
		}
	} else if policy == "FIFO" {

		_cache.Set(extraKey, "extra_value", NoExpiration)

		// 检查是否按预期淘汰了键
		_, found := _cache.Get(firstKey)
		if found {
			t.Errorf("FIFO eviction failed: %s should have been evicted in segment %d", firstKey, segmentIndexOfExtraKey)
		}
	}

}

func BenchmarkCacheGetManyConcurrent(b *testing.B) {
	b.StopTimer()

	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   1024,
		MaxCacheSize:   10000,
		EvictionPolicy: "LRU",
	})

	// Load the go-cache library for comparison
	//_cache := cache.New(5*time.Minute, 0)

	totalHits := int64(0)

	for i := 0; i < 1000000; i++ {
		_cache.Set(fmt.Sprintf("key%d", i), "value"+strconv.Itoa(i), NoExpiration)
	}

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			localHits := int64(0)
			for j := 0; j < each; j++ {
				x := rand.Intn(10000)
				key := fmt.Sprintf("key%d", x) // Random access to generate a certain hit rate
				value, found := _cache.Get(key)
				if found && value == "value"+strconv.Itoa(x) {
					localHits++
				}
			}
			atomic.AddInt64(&totalHits, localHits)
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()

	hitRate := float64(totalHits) / float64(b.N)
	b.Logf("Total items: %d, Each: %d, Hit Rate: %.2f%%", _cache.ItemCount(), each, hitRate*100)
}

func BenchmarkMapGetManyConcurrent(b *testing.B) {
	b.StopTimer()

	totalHits := int64(0)

	m := map[string]string{}
	mu := sync.RWMutex{}

	for i := 0; i < 1000000; i++ {
		mu.Lock()
		m[fmt.Sprintf("key%d", i)] = "value" + strconv.Itoa(i)
		mu.Unlock()
	}

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			localHits := int64(0)
			for j := 0; j < each; j++ {
				x := rand.Intn(10000)
				key := fmt.Sprintf("key%d", x) // Random access to generate a certain hit rate
				mu.RLock()
				value, found := m[key]
				mu.RUnlock()
				if found && value == "value"+strconv.Itoa(x) {
					localHits++
				}
			}
			atomic.AddInt64(&totalHits, localHits)
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()

	hitRate := float64(totalHits) / float64(b.N)
	b.Logf("Total items: %d, Each: %d, Hit Rate: %.2f%%", len(m), each, hitRate*100)
}

func BenchmarkCacheSetManyConcurrent(b *testing.B) {
	b.StopTimer()

	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   1024,
		MaxCacheSize:   10000,
		EvictionPolicy: "LRU",
	})

	// Load the go-cache library for comparison
	//_cache := cache.New(5*time.Minute, 0)

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			for j := 0; j < each; j++ {
				key := fmt.Sprintf("key%d", j)
				_cache.Set(key, "value"+strconv.Itoa(j), NoExpiration)
			}
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()
}

func BenchmarkMapSetManyConcurrent(b *testing.B) {
	b.StopTimer()

	var m = make(map[string]string)
	var mu sync.RWMutex

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			for j := 0; j < each; j++ {
				key := fmt.Sprintf("key%d", j)
				mu.Lock()
				m[key] = "value" + strconv.Itoa(j)
				mu.Unlock()
			}
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()
}

func BenchmarkCacheSetAndGetManyConcurrent(b *testing.B) {
	b.StopTimer()

	_cache, _ := NewCache(CacheConfig{
		SegmentCount:   1024,
		MaxCacheSize:   10000,
		EvictionPolicy: "LRU",
	})

	// Load the go-cache library for comparison
	//_cache := cache.New(5*time.Minute, 0)

	totalHits := int64(0) // 用于统计命中次数

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			localHits := int64(0)
			for j := 0; j < each; j++ {
				key := fmt.Sprintf("key%d", j)
				expectedValue := "value" + strconv.Itoa(j)
				_cache.Set(key, expectedValue, NoExpiration)
				actualValue, found := _cache.Get(key)
				if found && actualValue == expectedValue {
					localHits++
				}
			}
			atomic.AddInt64(&totalHits, localHits)
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()

	hitRate := float64(totalHits) / float64(b.N)
	b.Logf("Total items: %d, Each: %d, Hit Rate: %.2f%%", _cache.ItemCount(), each, hitRate*100)
}

func BenchmarkMapSetAndGetManyConcurrent(b *testing.B) {
	b.StopTimer()

	var m = make(map[string]string)
	var mu sync.RWMutex
	totalHits := int64(0)

	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			localHits := int64(0)
			for j := 0; j < each; j++ {
				key := fmt.Sprintf("key%d", j)
				expectedValue := "value" + strconv.Itoa(j)

				mu.Lock()
				m[key] = expectedValue
				mu.Unlock()

				// Read from the map
				mu.RLock()
				actualValue, found := m[key]
				mu.RUnlock()

				if found && actualValue == expectedValue {
					localHits++
				}
			}
			atomic.AddInt64(&totalHits, localHits)
			wg.Done()
		}(i)
	}

	b.StartTimer()
	wg.Wait()
	b.StopTimer()

	hitRate := float64(totalHits) / float64(b.N)
	b.Logf("Total operations: %d, Hit Rate: %.2f%%", b.N, hitRate*100)
}

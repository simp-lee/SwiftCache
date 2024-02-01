# SwiftCache

`SwiftCache` is a streamlined, in-memory cache library for Go, inspired by the **segmented caching concepts** from [freecache](https://github.com/coocood/freecache) and [bigcache](https://github.com/allegro/bigcache). Optimized for single-machine use, it encompasses under **600 lines** of code including comments, offering a robust yet concise caching solution. Influenced by [go-cache](https://github.com/patrickmn/go-cache), `SwiftCache` features a user-friendly interface for effortless integration.

A key feature is its thread-safe `map[string]interface{}` structure with support for entry expiration, eliminating the need for content serialization or network transmission. This allows for versatile object storage within the cache.

SwiftCache stands out with both Least Recently Used (**LRU**) and First In First Out (**FIFO**) eviction policies, giving developers control over data eviction according to their application's requirements. LRU prioritizes evicting least-accessed items, while FIFO removes items based on their addition order, catering to varied caching strategies.

## Performance

SwiftCache has undergone extensive benchmarking against traditional Go maps and the widely-used go-cache library, showcasing its superior performance in scenarios characterized by high concurrency and large data volumes. The following benchmarks highlight SwiftCache's efficiency, especially when managing **millions of items** and performing concurrent operations. 

Benchmarks source code can be found [here](https://github.com/simp-lee/SwiftCache/blob/main/cache_test.go#L885).

```
goos: windows
goarch: amd64
pkg: github.com/simp-lee/swiftcache
cpu: 13th Gen Intel(R) Core(TM) i5-13500H
BenchmarkGetManyConcurrent-16           27213105                41.54 ns/op
--- BENCH: BenchmarkGetManyConcurrent-16
cache_test.go:932: Total items: 1000000, Each: 0, Hit Rate: 0.00%
cache_test.go:932: Total items: 1000000, Each: 6, Hit Rate: 96.00%
cache_test.go:932: Total items: 1000000, Each: 625, Hit Rate: 100.00%
cache_test.go:932: Total items: 1000000, Each: 62500, Hit Rate: 100.00%
cache_test.go:932: Total items: 1000000, Each: 1700819, Hit Rate: 100.00%
...
```

The following benchmark summary showcases SwiftCache's performance in different scenarios:

```
+--------------------------+--------------------+----------------+-----------------+----------+
| Benchmark Scenario       | Cache Solution     | Operations/sec | Latency (ns/op) | Hit Rate |
+--------------------------+--------------------+----------------+-----------------+----------+
| GetManyConcurrent        | SwiftCache FIFO    | 27,213,105     | 41.54           | 100.00%  |
|                          | SwiftCache LRU     | 32,707,440     | 50.37           | 100.00%  |
|                          | go-cache           | 18,846,588     | 66.29           | 100.00%  |
|                          | Map                | 21,343,122     | 54.70           | 100.00%  |
+--------------------------+--------------------+----------------+-----------------+----------+
| SetManyConcurrent        | SwiftCache FIFO    | 15,027,386     | 88.20           | -        |
|                          | SwiftCache LRU     | 14,252,052     | 85.15           | -        |
|                          | go-cache           | 3,881,776      | 357.3           | -        |
|                          | Map                | 5,064,476      | 283.7           | -        |
+--------------------------+--------------------+----------------+-----------------+----------+
| SetAndGetManyConcurrent  | SwiftCache FIFO    | 10,798,873     | 95.76           | 100.00%  |
|                          | SwiftCache LRU     | 11,510,503     | 103.2           | 100.00%  |
|                          | go-cache           | 3,054,723      | 412.2           | 100.00%  |
|                          | Map                | 3,459,856      | 354.6           | 100.00%  |
+--------------------------+--------------------+----------------+-----------------+----------+

```
**GetManyConcurrent Test:** SwiftCache demonstrates exceptional retrieval speeds, with FIFO mode slightly outperforming LRU due to its simpler eviction logic. Both exhibit performance marginally superior to standard Go maps and `go-cache`, consistently achieving up to a 100% hit rate.

**SetManyConcurrent Test:** In write-heavy scenarios, SwiftCache maintains high performance, with FIFO mode again showing a slight edge over LRU. The efficiency starkly contrasts with the higher latencies observed in `go-cache` and Go maps.

**SetAndGetManyConcurrent Test:** When testing combined set and get operations, SwiftCache's efficiency shines, offering the lowest operation times among the tested solutions and maintaining a 100% hit rate.

## Usage

### Installation

```bash
go get github.com/simp-lee/swiftcache
```

### Simple initialization

A quick start guide to using SwiftCache:

```go
package main

import (
	"fmt"
	"github.com/simp-lee/swiftcache"
)

func main() {
	// Initialize the cache with default settings
	c, _ := swiftcache.NewCache()

	// Set a value with default expiration
	c.Set("myKey", "myValue", swiftcache.DefaultExpiration)

	// Retrieve and check if the value exists
	value, found := c.Get("myKey")
	if found {
		fmt.Println("Found myKey: ", value)
	}
}

```

### Custom initialization

To customize SwiftCache's behavior, such as segment count or eviction policy:

```go
package main

import (
	"github.com/simp-lee/swiftcache"
	"hash/fnv"
	"time"
)

func main() {
	cacheConfig := swiftcache.CacheConfig{
		SegmentCount:      1024,            // Customize the number of segments
		MaxCacheSize:      5000,            // Set the maximum cache size
		DefaultExpiration: 5 * time.Minute, // Default expiration time
		EvictionPolicy:    "LRU",           // Set the eviction policy
		HashFunc:          fnv.New32,       // Use FNV-1 32-bit hash function
	}

	// Initialize the cache with custom settings
	c, _ := swiftcache.NewCache(cacheConfig)

	// Use the cache with the configured settings
	// ...
}
```

## How it works

### Segmented Storage Mechanism

SwiftCache utilizes a segmented storage mechanism to distribute cache data across multiple segments or "shards". This approach significantly reduces lock contention, a common bottleneck in traditional, single-lock cache systems. By partitioning the cache space, SwiftCache allows concurrent access to different segments, enhancing throughput and scalability. Each segment operates independently with its own lock, ensuring that operations on different segments do not block each other.

### Superior Performance

The performance benefits of SwiftCache over standard Go maps and the `go-cache` library stem from this segmented architecture and the tailored locking strategy. Go maps, while fast for single-threaded operations, can suffer from contention in concurrent scenarios. SwiftCache's design minimizes this contention, providing faster access and modification times in multi-threaded environments.

### LRU and FIFO Eviction Policies

SwiftCache implements two eviction policies: Least Recently Used (LRU) and First In First Out (FIFO).

**LRU**: Items accessed least recently are evicted first. This policy is ideal for retaining frequently accessed data in the cache. Implementing LRU requires updating access order on each retrieval, which slightly affects performance due to the necessary lock operations for order maintenance.

**FIFO**: Items are evicted in the order they were added, without considering access patterns. FIFO simplifies the eviction process as it does not require updating order on access, leading to potentially higher performance for scenarios where access order is not critical.

The choice between FIFO and LRU affects the locking mechanism used. LRU's need for order updates requires more sophisticated locking strategies, whereas FIFO can often operate with simpler, more performance-optimized locking due to its straightforward eviction approach.

### Lazy Eviction and Expiration

Data eviction and expiration in SwiftCache are handled lazily. Instead of performing periodic sweeps to clean expired or evictable items, these operations are triggered during access attempts. This strategy ensures that the overhead of cleaning is spread out over time, preventing spikes in processing time that can occur with batch eviction or expiration processes. Lazy eviction contributes to a smoother performance profile, effectively "smoothing out" potential performance peaks.

### Bitwise Operations for Segment Allocation

SwiftCache optimizes key allocation to segments using bitwise AND operations, leveraging the requirement for the number of segments to be powers of two (e.g., 2, 4, 8, 16, ... 512, 1024). This design allows SwiftCache to replace conventional modulo arithmetic with a faster bitwise operation. Bitwise AND is used because, for any number of segments that is a power of two, calculating hash & (numberOfSegments - 1) effectively performs a modulo operation but with improved performance. Although the gain from this optimization is relatively minor, it underscores SwiftCache's dedication to maximizing efficiency across all aspects of its architecture. This approach not only simplifies the internal logic for segment allocation but also contributes to the overall performance advantages of SwiftCache in high-concurrency scenarios.

In summary, SwiftCache's design principles—segmented storage, tailored eviction policies, lazy data management, and optimized segment allocation—work in concert to provide a high-performance caching solution suitable for a wide range of applications.

## API

SwiftCache provides a simple yet powerful API, including methods for setting, getting, and deleting cache entries, as well as incrementing and decrementing numerical values. For a full list of methods and their descriptions, see our API documentation.

## Acknowledgments

Special thanks to the creators of [freecache](https://github.com/coocood/freecache) and [bigcache](https://github.com/allegro/bigcache) for their innovative caching mechanisms in Go, which significantly inspired the design of SwiftCache.

A heartfelt acknowledgment to [go-cache](https://github.com/patrickmn/go-cache) and its contributors. The design of SwiftCache's API and a substantial part of our testing methodologies are influenced by the simplicity and effectiveness of go-cache. Its approach to caching solutions has been crucial in guiding the development of a user-friendly interface and robust functionality in SwiftCache.

We are thankful to the open-source community for the collaboration and knowledge-sharing that make projects like SwiftCache possible. It is through these community efforts that we can continue to build efficient and effective software solutions.
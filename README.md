# SwiftCache

SwiftCache is a high-efficiency, in-memory cache library for Go applications, inspired by the segmented caching concepts from [freecache](https://github.com/coocood/freecache) and [bigcache](https://github.com/allegro/bigcache). Streamlined and optimized for single-machine use, the entire library, complete with detailed comments, is under 600 lines, making it a compact yet powerful caching solution. 

In drawing inspiration from the [go-cache](https://github.com/patrickmn/go-cache) library,SwiftCache has been designed with a user-friendly interface, ensuring ease of use right from the get-go.

One of the major advantages of SwiftCache is its core functionality as a thread-safe `map[string]interface{}` that supports expiration times for its entries. This design choice means that there is no need for the contents of the cache to be serialized or transmitted over the network, simplifying the caching process considerably. Any object can be stored.

Moreover, SwiftCache is distinguished by its implementation of both Least Recently Used (LRU) and First In First Out (FIFO) eviction policies. These eviction strategies offer developers the flexibility to choose how cached data is managed and purged, based on their specific application needs. With LRU, items that have not been accessed recently are the first to be evicted, making room for new or more frequently accessed data. On the other hand, the FIFO strategy evicts items in the order they were added, regardless of access frequency.

## Performance

Benchmarking against traditional Go maps and similar cache solutions demonstrates SwiftCache's exceptional performance, especially in environments with high concurrency and large datasets. Detailed benchmark results and comparisons are available in our [performance documentation](#).

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
    "time"
)

func main() {
    cacheConfig := swiftcache.CacheConfig{
        SegmentCount: 1024, // Customize the number of segments
        MaxCacheSize: 5000, // Set the maximum cache size
        DefaultExpiration: 5 * time.Minute, // Default expiration time
        EvictionPolicy: "LRU", // Set the eviction policy
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
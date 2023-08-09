package shardmap

import (
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	ShardCount = 32
	offset32   = 2166136261
	prime32    = 16777619
)

// ShardMap is a map sharded into multiple concurrent maps for better parallelism.
type ShardMap struct {
	shards []*sync.Map
}

// NewShardMap creates and initializes a new ShardMap with multiple shards.
func New() *ShardMap {
	s := &ShardMap{
		shards: make([]*sync.Map, ShardCount),
	}
	for i := 0; i < ShardCount; i++ {
		s.shards[i] = &sync.Map{}
	}
	return s
}

// getShard returns the appropriate shard for the given key.
func (s *ShardMap) getShard(key any) *sync.Map {
	index := uint(fnv1a32(key)) % uint(ShardCount)
	return s.shards[index]
}

// Load retrieves the value associated with the key from the ShardMap.
func (s *ShardMap) Load(key any) (value any, ok bool) {
	shard := s.getShard(key)
	value, ok = shard.Load(key)
	return
}

// Store stores the key-value pair in the ShardMap.
func (s *ShardMap) Store(key any, value any) {
	shard := s.getShard(key)
	shard.Store(key, value)
}

// Delete removes the key and its associated value from the ShardMap.
func (s *ShardMap) Delete(key any) {
	shard := s.getShard(key)
	shard.Delete(key)
}

// Len returns the total number of key-value pairs in the ShardMap.
func (s *ShardMap) Len() uint32 {
	var count uint32
	wg := sync.WaitGroup{}
	wg.Add(ShardCount)

	// Iterate over all shards concurrently to count key-value pairs.
	for index, shard := range s.shards {
		go func(i int, c *sync.Map) {
			c.Range(func(key, value any) bool {
				// Increment count atomically to avoid race conditions.
				atomic.AddUint32(&count, 1)
				return true
			})
			wg.Done()
		}(index, shard)
	}
	wg.Wait()
	return count
}

// Range iterates over all key-value pairs in the ShardMap and applies the given function.
func (s *ShardMap) Range(fn func(key, value any)) {
	wg := sync.WaitGroup{}
	wg.Add(ShardCount)

	// Iterate over all shards concurrently and apply the given function to each key-value pair.
	for index, shard := range s.shards {
		go func(i int, c *sync.Map, f func(key, value any)) {
			c.Range(func(key, value any) bool {
				// Apply the provided function to the key-value pair.
				f(key, value)
				return true
			})
			wg.Done()
		}(index, shard, fn)
	}
	wg.Wait()
}

// Clear removes all key-value pairs from the ShardMap.
func (s *ShardMap) Clear() {
	wg := sync.WaitGroup{}
	wg.Add(ShardCount)

	// Iterate over all shards concurrently to clear all key-value pairs.
	for index, shard := range s.shards {
		go func(i int, c *sync.Map) {
			c.Range(func(key, value any) bool {
				// Delete each key-value pair from the shard.
				c.Delete(key)
				return true
			})
			wg.Done()
		}(index, shard)
	}
	wg.Wait()
}

// fnv1a32 calculates the FNV-1a hash value for the given key,
// supporting various types including integers, strings, and pointers.
func fnv1a32(key any) uint32 {
	var strKey string

	// Convert the key to a string for hash calculation based on its type.
	switch v := key.(type) {
	case string:
		strKey = v
	case int:
		strKey = strconv.Itoa(v)
	case int8:
		strKey = strconv.Itoa(int(v))
	case int16:
		strKey = strconv.Itoa(int(v))
	case int32:
		strKey = strconv.Itoa(int(v))
	case int64:
		strKey = strconv.Itoa(int(v))
	case uint:
		strKey = strconv.FormatUint(uint64(v), 10)
	case uint8:
		strKey = strconv.FormatUint(uint64(v), 10)
	case uint16:
		strKey = strconv.FormatUint(uint64(v), 10)
	case uint32:
		strKey = strconv.FormatUint(uint64(v), 10)
	case uint64:
		strKey = strconv.FormatUint(v, 10)
	case uintptr:
		strKey = strconv.FormatUint(uint64(v), 10)
	default:
		// Unsupported key type, panic
		panic("Unsupported key type")
	}

	// Calculate hash value using the FNV-1a hash algorithm.
	hash := uint32(offset32)
	keyLength := len(strKey)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(strKey[i])
	}
	return hash
}

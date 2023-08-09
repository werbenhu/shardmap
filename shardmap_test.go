package shardmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New()
	assert.Equal(t, ShardCount, len(s.shards), "Expected %d shards, got %d", ShardCount, len(s.shards))
}

func TestLoadStoreDelete(t *testing.T) {
	s := New()

	// Test Store and Load
	s.Store("key1", "value1")
	value, ok := s.Load("key1")
	assert.True(t, ok, "Load/Store test failed")
	assert.Equal(t, "value1", value, "Load/Store test failed")

	// Test Delete
	s.Delete("key1")
	_, ok = s.Load("key1")
	assert.False(t, ok, "Delete test failed")
}

func TestLen(t *testing.T) {
	s := New()
	s.Store("key1", "value1")
	s.Store("key2", "value2")

	length := s.Len()
	assert.Equal(t, uint32(2), length, "Expected length 2, got %d", length)
}

func TestRange(t *testing.T) {
	s := New()
	s.Store("key1", "value1")
	s.Store("key2", "value2")

	// Define the test function for Range
	testFunc := func(key, value any) {
		if key.(string) == "key1" {
			s.Delete(key)
		}
	}

	// Run the Range function
	s.Range(testFunc)

	// Check values
	_, ok := s.Load("key1")
	assert.False(t, ok, "Range test failed")

	_, ok = s.Load("key2")
	assert.True(t, ok, "Range test failed")

	// Check len after range
	length := s.Len()
	assert.Equal(t, uint32(1), length, "Range test len failed after clear")
}

func TestClear(t *testing.T) {
	s := New()
	s.Store("key1", "value1")
	s.Store("key2", "value2")

	s.Clear()

	// Check if all values are cleared
	_, ok := s.Load("key1")
	assert.False(t, ok, "Clear test failed")

	_, ok = s.Load("key2")
	assert.False(t, ok, "Clear test failed")

	// Check len after cleared
	length := s.Len()
	assert.Equal(t, uint32(0), length, "Clear test len failed after clear")
}

func TestFnv1a32(t *testing.T) {
	// Test with different key types
	tests := []struct {
		key      interface{}
		expected uint32
	}{
		{"test_string", 2101758991},
		{42, 494316163},
	}

	for _, test := range tests {
		result := fnv1a32(test.key)
		assert.Equal(t, test.expected, result, "Expected hash %d for key %v, got %d", test.expected, test.key, result)
	}
}

func TestConcurrentLoadPanic(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		m.Store(i, i)
	}

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, ok := m.Load(j)
				assert.True(t, ok, "ConcurrentLoadPanic test failed: Load failed for key %d", j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestConcurrentStorePanic(t *testing.T) {
	m := New()

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.Store(j, j)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.NotPanics(t, func() {
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					_, ok := m.Load(j)
					assert.True(t, ok, "ConcurrentStorePanic test failed: Load failed for key %d", j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestStoreOrLoadConcurrent(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		m.Store(i, i)
	}

	storewg := sync.WaitGroup{}
	storewg.Add(100)
	assert.NotPanics(t, func() {
		for i := 0; i < 100; i++ {
			go func(index int) {
				for j := index * 100; j < (index+1)*100; j++ {
					m.Store(j, j)
				}
				storewg.Done()
			}(i)
		}
	})
	storewg.Wait()

	loadwg := sync.WaitGroup{}
	loadwg.Add(100)
	assert.NotPanics(t, func() {
		for i := 0; i < 100; i++ {
			go func(index int) {
				for j := index * 100; j < (index+1)*100; j++ {
					val, ok := m.Load(j)
					assert.True(t, ok, "StoreOrLoadConcurrent test failed: Load failed for key %d", j)
					assert.Equal(t, j, val, "StoreOrLoadConcurrent test failed: Unexpected value for key %d", j)
				}
				loadwg.Done()
			}(i)
		}
	})
	loadwg.Wait()
}

func TestStoreAndLoadConcurrent(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		m.Store(i, i)
	}

	assert.NotPanics(t, func() {
		loadGoroutineSize := 100
		loadWg := sync.WaitGroup{}
		loadWg.Add(loadGoroutineSize)

		for i := 0; i < loadGoroutineSize; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					val, ok := m.Load(j)
					assert.True(t, ok, "StoreAndLoadConcurrent test failed: Load failed for key %d", j)
					assert.Equal(t, j, val, "StoreAndLoadConcurrent test failed: Unexpected value for key %d", j)
				}
				loadWg.Done()
			}()
		}

		storeGoroutineSize := 100
		storeWg := sync.WaitGroup{}
		storeWg.Add(storeGoroutineSize)
		for i := 0; i < storeGoroutineSize; i++ {
			go func(index int) {
				for j := 0; j < 100; j++ {
					m.Store(j, j)
				}
				storeWg.Done()
			}(i)
		}

		storeWg.Wait()
		loadWg.Wait()
	})
}

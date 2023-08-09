package shardmap

import (
	"strconv"
	"sync"
	"testing"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type imap struct {
	create func() interface{}
	store  func(m interface{}, k, v string)
	load   func(m interface{}, k string) (any, bool)
}

func getMaps(single bool) map[string]*imap {

	maps := make(map[string]*imap)
	if single {
		maps["map"] = &imap{
			create: func() interface{} { return make(map[string]string) },
			store:  func(m interface{}, k, v string) { m.(map[string]string)[k] = v },
			load: func(m interface{}, k string) (any, bool) {
				k, ok := m.(map[string]string)[k]
				return k, ok
			},
		}
	}

	maps["ShardMap"] = &imap{
		create: func() interface{} { return New() },
		store:  func(m interface{}, k, v string) { m.(*ShardMap).Store(k, v) },
		load: func(m interface{}, k string) (any, bool) {
			return m.(*ShardMap).Load(k)
		},
	}

	maps["concurrent-map"] = &imap{
		create: func() interface{} { return cmap.New[string]() },
		store:  func(m interface{}, k, v string) { m.(cmap.ConcurrentMap[string, string]).Set(k, v) },
		load: func(m interface{}, k string) (any, bool) {
			return m.(cmap.ConcurrentMap[string, string]).Get(k)
		},
	}

	maps["sync.Map"] = &imap{
		create: func() interface{} { return &sync.Map{} },
		store:  func(m interface{}, k, v string) { m.(*sync.Map).Store(k, v) },
		load: func(m interface{}, k string) (any, bool) {
			return m.(*sync.Map).Load(k)
		},
	}

	return maps
}

// BenchmarkSingleGoroutineStoreAbsent tests storing absent keys using different map types.
func BenchmarkSingleGoroutineStoreAbsent(b *testing.B) {
	maps := getMaps(true)

	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.store(m, strconv.Itoa(i), "value")
			}
			b.StopTimer()
		})
	}
}

// BenchmarkSingleGoroutineStoreAbsent tests storing absent keys using different map types.
func BenchmarkSingleGoroutineStorePresent(b *testing.B) {
	maps := getMaps(true)
	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.store(m, "key", "value")
			}
			b.StopTimer()
		})
	}
}

func BenchmarkMultiGoroutineStoreDifferent(b *testing.B) {
	maps := getMaps(false)
	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()
			storeNum := 10000
			finished := make(chan struct{}, b.N)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				go func(k int) {
					for j := 0; j < storeNum; j++ {
						val := strconv.Itoa(k + j)
						bm.store(m, val, val)
					}
					finished <- struct{}{}
				}(i)
			}
			for i := 0; i < b.N; i++ {
				<-finished
			}
		})
	}
}

func BenchmarkMultiGoroutineLoadSame(b *testing.B) {
	maps := getMaps(false)
	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()

			bm.store(m, "key", "value")
			finished := make(chan struct{}, b.N)
			b.ResetTimer()
			loadNum := 1000
			for i := 0; i < b.N; i++ {
				go func(k int) {
					for j := 0; j < loadNum; j++ {
						bm.load(m, "key")
					}
					finished <- struct{}{}
				}(i)
			}
			for i := 0; i < b.N; i++ {
				<-finished
			}
		})
	}
}

func BenchmarkMultiGoroutineLoadDifferent(b *testing.B) {
	maps := getMaps(false)

	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()

			storeNum := 100000
			for i := 0; i < storeNum; i++ {
				val := strconv.Itoa(i)
				bm.store(m, val, val)
			}

			finished := make(chan struct{}, b.N)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				go func(k int) {
					for j := 0; j < storeNum; j++ {
						bm.load(m, strconv.Itoa(j))
					}
					finished <- struct{}{}
				}(i)
			}
			for i := 0; i < b.N; i++ {
				<-finished
			}
		})
	}
}

func BenchmarkMultiGoroutineStoreAndLoadDifferent(b *testing.B) {
	maps := getMaps(false)

	for name, bm := range maps {
		b.Run(name, func(b *testing.B) {
			m := bm.create()

			storeNum := 10000
			for i := 0; i < storeNum; i++ {
				val := strconv.Itoa(i)
				bm.store(m, val, val)
			}

			finished := make(chan struct{}, 2*b.N)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				go func(k int) {
					for j := 0; j < storeNum; j++ {
						val := strconv.Itoa(i)
						bm.store(m, val, val)
					}
					finished <- struct{}{}
				}(i)
			}

			for i := 0; i < b.N; i++ {
				go func(k int) {
					for i := 0; i < storeNum; i++ {
						bm.load(m, strconv.Itoa(i))
					}
					finished <- struct{}{}
				}(i)
			}
			for i := 0; i < 2*b.N; i++ {
				<-finished
			}
		})
	}
}

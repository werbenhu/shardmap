# shardmap

## a thread-safe concurrent map for go
## 一个线程安全、支持并发的map。

`shardmap` 结合了 [concurrent-map](https://github.com/orcaman/concurrent-map) 和 [sync.Map](https://github.com/golang/go/tree/master/src/sync) 各自的优点, 在sync.Map的基础上进行分片，使其具备更好的写入性能。

## 压测
下面是压力测试的结果
```
cpu: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz
BenchmarkSingleGoroutineStoreAbsent/map-4                           2765622     370.9 ns/op
BenchmarkSingleGoroutineStoreAbsent/ShardMap-4                      1000000     1065 ns/op
BenchmarkSingleGoroutineStoreAbsent/concurrent-map-4                2787558     409.3 ns/op
BenchmarkSingleGoroutineStoreAbsent/sync.Map-4                      1000000     1076 ns/op

BenchmarkSingleGoroutineStorePresent/map-4                          74137227    15.66 ns/op
BenchmarkSingleGoroutineStorePresent/ShardMap-4                     6316366     185.7 ns/op
BenchmarkSingleGoroutineStorePresent/concurrent-map-4               22964001    48.04 ns/op
BenchmarkSingleGoroutineStorePresent/sync.Map-4                     7025263     175.7 ns/op

BenchmarkMultiGoroutineStoreDifferent/ShardMap-4                    870         1289748 ns/op
BenchmarkMultiGoroutineStoreDifferent/concurrent-map-4              1662        677558 ns/op
BenchmarkMultiGoroutineStoreDifferent/sync.Map-4                    348         3483560 ns/op

BenchmarkMultiGoroutineLoadSame/ShardMap-4                          121842      9145 ns/op
BenchmarkMultiGoroutineLoadSame/concurrent-map-4                    16670       68927 ns/op
BenchmarkMultiGoroutineLoadSame/sync.Map-4                          145180      9004 ns/op

BenchmarkMultiGoroutineLoadDifferent/ShardMap-4                     301         3694477 ns/op
BenchmarkMultiGoroutineLoadDifferent/concurrent-map-4               210         5809354 ns/op
BenchmarkMultiGoroutineLoadDifferent/sync.Map-4                     340         3263945 ns/op

BenchmarkMultiGoroutineStoreAndLoadDifferent/ShardMap-4             986         1199937 ns/op
BenchmarkMultiGoroutineStoreAndLoadDifferent/concurrent-map-4       404         3046184 ns/op
BenchmarkMultiGoroutineStoreAndLoadDifferent/sync.Map-4             1014        1228813 ns/op
```

## 使用

引入包:
```
import (
	"github.com/werbenhu/shardmap"
)
```
go get "github.com/werbenhu/shardmap"

## 示例

```
package main

import (
	"fmt"
	"strconv"

	"github.com/werbenhu/shardmap"
)

func main() {
	sm := shardmap.New()

	for i := 0; i < 10; i++ {
		sm.Store("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
	}

	fmt.Printf("Len:%d\n", sm.Len())

	val, ok := sm.Load("key8")
	if ok {
		fmt.Printf("Load key8 value:%s\n", val)
	}

	sm.Range(func(key, value any) {
		fmt.Printf("Rang key:%s value:%s\n", key, value)
	})

	sm.Clear()
	fmt.Printf("After clear Len:%d\n", sm.Len())
}

```
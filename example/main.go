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

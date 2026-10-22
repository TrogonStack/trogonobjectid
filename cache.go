package trogonobjectid

import (
	"reflect"
	"sync"
)

type cached[V any] struct {
	value V
	err   error
}

// loadOrBuild returns the value cached for key, building it on first use. The cache
// holds pointers, so a lookup on the encode and parse paths copies no struct.
func loadOrBuild[V any](cache *sync.Map, key reflect.Type, build func() (V, error)) (*V, error) {
	if c, ok := cache.Load(key); ok {
		entry := c.(*cached[V])
		return &entry.value, entry.err
	}
	value, err := build()
	c, _ := cache.LoadOrStore(key, &cached[V]{value: value, err: err})
	entry := c.(*cached[V])
	return &entry.value, entry.err
}

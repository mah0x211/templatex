package templatex

import "sync"

type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, v interface{})
	Unset(k string)
}

type NopCache struct {
	Cache
}

func NewNopCache() Cache {
	return NopCache{}
}

func (c NopCache) Get(_ string) (interface{}, bool) {
	return nil, false
}
func (c NopCache) Set(_ string, _ interface{}) {}
func (c NopCache) Unset(_ string)              {}

type MapCache struct {
	Cache
	sync.Mutex
	data map[string]interface{}
}

func NewMapCache() Cache {
	return &MapCache{
		data: make(map[string]interface{}),
	}
}

func (c *MapCache) Get(k string) (interface{}, bool) {
	c.Lock()
	v, ok := c.data[k]
	c.Unlock()
	return v, ok
}

func (c *MapCache) Set(k string, v interface{}) {
	c.Lock()
	c.data[k] = v
	c.Unlock()
}

func (c *MapCache) Unset(k string) {
	c.Lock()
	delete(c.data, k)
	c.Unlock()
}

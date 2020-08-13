package templatex

import "sync"

type File struct {
	cache  Cache
	name   string
	tmpl   interface{}
	parent map[string]*File
	child  map[string]*File
}

func createFile(cache Cache, name string) *File {
	return &File{
		cache:  cache,
		name:   name,
		parent: make(map[string]*File),
		child:  make(map[string]*File),
	}
}

func (f *File) Name() string {
	return f.name
}

func (f *File) addParent(af *File) {
	f.parent[af.name] = af
}

func (f *File) addChild(af *File) {
	f.child[af.name] = af
}

func (f *File) Uncache() {
	f.cache.Unset(f.name)
	for _, p := range f.parent {
		p.Uncache()
	}
}

type Cache interface {
	Get(k string) *File
	Set(k string, f *File)
	Unset(k string)
}

type NopCache struct {
	Cache
}

func NewNopCache() Cache {
	return NopCache{}
}

func (c NopCache) Get(_ string) *File {
	return nil
}
func (c NopCache) Set(_ string, _ *File) {}
func (c NopCache) Unset(_ string)        {}

type MapCache struct {
	Cache
	sync.Mutex
	data map[string]*File
}

func NewMapCache() Cache {
	return &MapCache{
		data: make(map[string]*File),
	}
}

func (c *MapCache) Get(k string) *File {
	c.Lock()
	defer c.Unlock()
	return c.data[k]
}

func (c *MapCache) Set(k string, f *File) {
	c.Lock()
	c.data[k] = f
	c.Unlock()
}

func (c *MapCache) Unset(k string) {
	c.Lock()
	delete(c.data, k)
	c.Unlock()
}

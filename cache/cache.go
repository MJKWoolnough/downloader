package cache

import (
	"path"
	"sync"

	"github.com/MJKWoolnough/downloader"
)

type Cache struct {
	objects map[string]*object
	mutex   sync.Mutex
	dir     string
}

func NewCache(dir string) *Cache {
	return &Cache{
		objects: make(map[string]*object),
		dir:     dir,
	}
}

func (c *Cache) Get(key string, r downloader.Downloader) (*CachedObject, error) {
	var err error
	c.mutex.Lock()
	defer c.mutex.Unlock()
	o, ok := c.objects[key]
	if !ok {
		o, err = newObject(path.Join(c.dir, key), r)
		if err != nil {
			return nil, err
		}
		c.objects[key] = o
	}
	return &CachedObject{
		o: o,
	}, nil
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if o, ok := c.objects[key]; ok {
		close(o.quit)
		delete(c.objects, key)
	}
}

func (c *Cache) Keys() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	os := make([]string, 0, len(c.objects))
	for key := range c.objects {
		os = append(os, key)
	}
	return os
}

func (c *Cache) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, o := range c.objects {
		close(o.quit)
		delete(c.objects, key)
	}
	return nil
}

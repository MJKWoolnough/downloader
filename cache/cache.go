package cache

import (
	"io"
	"os"
	"sync"
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

func LoadCache(dir string, os []string) *Cache {
	c := &Cache{
		objects: make(map[string]*object),
		dir:     dir,
	}
	for _, o := range os {
		c.objects[o] = &object{}
	}
	return c
}

func (c *Cache) Get(key string, r io.ReadCloser, start, length int) *CachedObject {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	o, ok := c.objects[key]
	if !ok {
		o := newObject(c, key, r)
		c.objects[key] = o
	}
	return &CachedObject{
		object:   o,
		filename: key,
		start:    start,
		length:   length,
	}
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if o, ok := c.objects[key]; ok {
		delete(c.objects, key)
		close(o.q)
	}
	return nil
}

func (c *Cache) Dir() string {
	return c.dir
}

func (c *Cache) Save() []string {
	os := make([]string, 0, len(c.objects))
	for key := range c.objects {
		os = append(os, key)
	}
	return os
}

func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.objects = make(map[string]*object)
	return os.RemoveAll(c.dir)
}

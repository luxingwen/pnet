package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GoCache struct {
	cache *gocache.Cache
}

func NewGoCache(defaultExpiration, cleanupInterval time.Duration) *GoCache {
	return &GoCache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}

func (gc *GoCache) byteKeyToStringKey(byteKey []byte) string {
	return string(byteKey)
}

func (gc *GoCache) Add(key []byte, value interface{}) error {
	return gc.cache.Add(gc.byteKeyToStringKey(key), value, 0)
}

func (gc *GoCache) AddWithExpiration(key []byte, value interface{}, expiration time.Duration) error {
	return gc.cache.Add(gc.byteKeyToStringKey(key), value, expiration)
}

func (gc *GoCache) Get(key []byte) (interface{}, bool) {
	return gc.cache.Get(gc.byteKeyToStringKey(key))
}

func (gc *GoCache) Set(key []byte, value interface{}) error {
	gc.cache.Set(gc.byteKeyToStringKey(key), value, 0)
	return nil
}

func (gc *GoCache) SetWithExpiration(key []byte, value interface{}, expiration time.Duration) error {
	gc.cache.Set(gc.byteKeyToStringKey(key), value, expiration)
	return nil
}

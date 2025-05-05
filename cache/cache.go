package cache

import(
	"sync"
	"time"
)

var cache sync.Map

func CacheItem(key string, value interface{}, expiration time.Duration) {
	cache.Store(key, value)
	time.AfterFunc(expiration, func() {
		cache.Delete(key)
	})
}

func IsCached(key string) (interface{}, bool) {
	return cache.Load(key)
}

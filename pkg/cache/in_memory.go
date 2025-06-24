package cache

import "github.com/patrickmn/go-cache"

type InMemoryCache struct {
	*cache.Cache
}

var inMemoryCacheInstance *InMemoryCache

func GetClient() *InMemoryCache {
	return inMemoryCacheInstance
}

func SetClient(cache *cache.Cache) {
	inMemoryCacheInstance = &InMemoryCache{cache}
}

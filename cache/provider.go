package cache

import (
	"time"
)

type Cacheable interface {
	CacheKey() string
}

type IDCacheable interface {
	GetID() int64
}

type CacheableObject interface {
	Cacheable
	GetCacheObject() map[string]string
	SetCacheObject(cacheData map[string]string) error
}

type CacheProvider interface {
	Test() error
	Exists(key string) (bool, error)
	Get(key string) (string, error)
	Save(key string, val any, ttl time.Duration) error
	Delete(key string) error
	TTL(key string) (time.Duration, error)
	UpdateTTl(key string, ttl time.Duration) (bool, error)

	SaveJSON(key string, val any, ttl time.Duration) error

	GetMap(key string) (map[string]string, error)
	SetMapValue(key string, field string, value string) (bool, error)
	GetMapValue(key string, field string) (string, error)
	MapKeyExists(key string, mapKey string) (bool, error)

	GetSet(key string) ([]string, error)
	AddSetMember(key string, member string) (bool, error)
	RemoveSetMember(key string, member string) (bool, error)

	SetPrefix(prefix string)
	GetPrefix() string
}

package cache

import (
	"time"
)

type Cacheable interface {
	CacheKey() string
	CacheTTL() time.Duration
}

type IDCacheable interface {
	GetID() int64
}

type CacheableObject interface {
	Cacheable
	SetCacheKey(key string)
	IsCacheLoaded() bool
	GetCacheObject() map[string]string
	SetCacheObject(cacheData map[string]string) error
}

type CacheProvider interface {
	Exists(key string) (bool, error)
	MustExists(key string) bool
	IsHSET(key string) bool

	Get(key string) (string, error)
	Save(key string, val any, ttl time.Duration) error
	Delete(key string) error
	TTL(key string) (time.Duration, error)
	UpdateTTl(key string, ttl time.Duration) (bool, error)

	SaveAsJSON(key string, val any, ttl time.Duration) error
	LoadJSON(key string, result any) (any, error)

	SaveObj(o CacheableObject) error
	LoadObj(o CacheableObject) error

	GetMap(key string) (map[string]string, error)
	GetMapKeys(key string, fields ...string) (map[string]string, error)
	GetMapValue(key string, field string) (string, error)
	MapKeyExists(key string, mapKey string) (bool, error)
	SetMapValue(key string, field string, value string) (bool, error)

	GetSet(key string) ([]string, error)
	AddSetMember(key string, members ...string) (bool, error)
	RemoveSetMember(key string, members ...string) (bool, error)

	SetPrefix(prefix string)
	GetPrefix() string

	Test() error
}

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/siper92/core-utils/config_utils"
	"time"
)

const InfiniteTTL = time.Duration(0)

var _ CacheProvider = (*RedisCacheProvider)(nil)

type RedisCacheProvider struct {
	Config config_utils.RedisConfig

	client *redis.Client
	ctx    context.Context

	prefix string
}

func NewClient(cnf config_utils.RedisConfig, prefix string) (*RedisCacheProvider, error) {
	var client *redis.Client

	client = redis.NewClient(&redis.Options{
		Addr:     cnf.GetAddr(),
		Password: cnf.Pass,
		DB:       cnf.Database,
	})

	cacheClient := &RedisCacheProvider{
		prefix: prefix,
		client: client,
		ctx:    context.Background(),
		Config: cnf,
	}

	err := cacheClient.Test()
	if err != nil {
		return nil, err
	}

	return cacheClient, nil
}

func (r *RedisCacheProvider) toKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

func (r *RedisCacheProvider) Client() *redis.Client {
	return r.client
}

func (r *RedisCacheProvider) Test() error {
	_, err := r.Client().Ping(r.ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %s", err)
	}

	return fmt.Errorf("implement me")
}

func (r *RedisCacheProvider) Exists(key string) (bool, error) {
	if r.Client().Exists(r.ctx, r.toKey(key)).Val() == 1 {
		return true, nil
	}

	return false, nil
}

func (r *RedisCacheProvider) Get(key string) (string, error) {
	return r.Client().Get(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) Save(key string, val any, ttl time.Duration) error {
	saveVal := "no data"
	switch valT := val.(type) {
	case CacheableObject:
		cacheData := valT.GetCacheObject()
		res := r.Client().HSet(r.ctx, r.toKey(key), cacheData, ttl)

		return res.Err()
	case string:
		saveVal = valT
	case []byte:
		saveVal = string(valT)
	case int, int32, int64, uint, uint32, uint64:
		saveVal = fmt.Sprintf("%d", val)
	case float32, float64:
		saveVal = fmt.Sprintf("%f", val)
	default:
		saveVal = fmt.Sprintf("%v", val)
	}

	return r.Client().Set(r.ctx, r.toKey(key), saveVal, ttl).Err()
}

func (r *RedisCacheProvider) Delete(key string) error {
	return r.Client().Del(r.ctx, r.toKey(key)).Err()
}

func (r *RedisCacheProvider) TTL(key string) (time.Duration, error) {
	return r.Client().TTL(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) UpdateTTl(key string, ttl time.Duration) (bool, error) {
	return r.Client().Expire(r.ctx, r.toKey(key), ttl).Result()
}

func (r *RedisCacheProvider) SaveJSON(key string, val any, ttl time.Duration) error {
	valJson, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return r.Save(r.toKey(key), valJson, ttl)
}

func (r *RedisCacheProvider) GetMap(key string) (map[string]string, error) {
	return r.Client().HGetAll(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) SetMapValue(key string, field string, value string) (bool, error) {
	if res := r.Client().HSet(r.ctx, r.toKey(key), field, value); res.Err() != nil {
		return false, res.Err()
	}

	return true, nil
}

func (r *RedisCacheProvider) GetMapValue(key string, field string) (string, error) {
	return r.Client().HGet(r.ctx, r.toKey(key), field).Result()
}

func (r *RedisCacheProvider) MapKeyExists(key string, mapKey string) (bool, error) {
	return r.Client().HExists(r.ctx, r.toKey(key), mapKey).Result()
}

func (r *RedisCacheProvider) GetSet(key string) ([]string, error) {
	return r.Client().SMembers(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) AddSetMember(key string, member string) (bool, error) {
	res, err := r.Client().SAdd(r.ctx, r.toKey(key), member).Result()
	if err != nil {
		return false, err
	}

	return res > 0, nil
}

func (r *RedisCacheProvider) RemoveSetMember(key string, member string) (bool, error) {
	res, err := r.Client().SRem(r.ctx, r.toKey(key), member).Result()
	if err != nil {
		return false, err
	}

	return res > 0, nil
}

func (r *RedisCacheProvider) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *RedisCacheProvider) GetPrefix() string {
	return r.prefix
}

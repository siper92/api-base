package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/siper92/core-utils/config_utils"
	"github.com/siper92/core-utils/type_utils"
	"reflect"
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

var defaultCacheClient *RedisCacheProvider

func Client() *RedisCacheProvider {
	if defaultCacheClient == nil || defaultCacheClient.client == nil {
		panic("redis cache provider not initialized")
	}

	return defaultCacheClient
}

func InitDefaultClient(cnf config_utils.RedisConfig, prefix string) error {
	var err error
	defaultCacheClient, err = NewClient(cnf, prefix)
	if err != nil {
		return err
	}

	Client()

	return nil
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

func (r *RedisCacheProvider) MustExists(key string) bool {
	ok, _ := r.Exists(key)
	return ok
}

func (r *RedisCacheProvider) IsHSET(key string) bool {
	res := r.client.Type(r.ctx, r.toKey(key)).Val()

	return res == "hash"

}

func (r *RedisCacheProvider) Get(key string) (string, error) {
	return r.Client().Get(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) Save(key string, val any, ttl time.Duration) error {
	// check if val is a map
	var saveVal any

	switch valT := val.(type) {
	case string:
		saveVal = valT
	case []byte:
		saveVal = string(valT)
	case int, int32, int64, uint, uint32, uint64:
		saveVal = fmt.Sprintf("%d", val)
	case float32, float64:
		saveVal = fmt.Sprintf("%f", val)
	default:
		if reflect.TypeOf(saveVal).Kind() == reflect.Map {
			res := r.Client().HMSet(r.ctx, r.toKey(key), val, ttl)

			return res.Err()
		} else if reflect.TypeOf(saveVal).Kind() == reflect.Slice {
			res, err := r.Client().SAdd(r.ctx, r.toKey(key), saveVal.([]any)...).Result()
			if err != nil {
				return err
			}

			if res == 0 && len(saveVal.([]any)) > 0 {
				return errors.New("failed to save set")
			}
		} else {
			return fmt.Errorf("unsupported type %T", val)
		}
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

func (r *RedisCacheProvider) SaveAsJSON(key string, val any, ttl time.Duration) error {
	valJson, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return r.Save(r.toKey(key), valJson, ttl)
}

func (r *RedisCacheProvider) LoadJSON(key string, result any) (any, error) {
	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return nil, errors.New("result argument must be a pointer")
	}

	val, err := r.Get(key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *RedisCacheProvider) GetMap(key string) (map[string]string, error) {
	return r.Client().HGetAll(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) GetMapKeys(key string, fields ...string) (map[string]string, error) {
	var resultMap map[string]string
	res := r.Client().HMGet(r.ctx, r.toKey(key), fields...)
	if res.Err() != nil {
		return resultMap, res.Err()
	}

	for i, v := range res.Val() {
		if v != nil {
			if _, ok := v.(string); !ok {
				resultMap[fields[i]] = v.(string)
			} else {
				resultMap[fields[i]] = type_utils.BaseTypeToString(v)
			}
		}
	}

	return resultMap, nil
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

func (r *RedisCacheProvider) AddSetMember(key string, members ...string) (bool, error) {
	membersInterfaces := make([]interface{}, len(members))
	for i, member := range members {
		membersInterfaces[i] = member
	}
	res, err := r.Client().SAdd(r.ctx, r.toKey(key), membersInterfaces...).Result()
	if err != nil {
		return false, err
	}

	return res > 0, nil
}

func (r *RedisCacheProvider) RemoveSetMember(key string, members ...string) (bool, error) {
	membersInterfaces := make([]interface{}, len(members))
	for i, member := range members {
		membersInterfaces[i] = member
	}
	res, err := r.Client().SRem(r.ctx, r.toKey(key), membersInterfaces...).Result()
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

func (r *RedisCacheProvider) LoadObj(obj CacheableObject) error {
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return fmt.Errorf("object %T must be a pointer", obj)
	}

	key := obj.CacheKey()
	if !r.MustExists(key) {
		return fmt.Errorf("cache key %s not found", key)
	}

	cacheData, err := r.GetMap(key)
	if err != nil {
		return err
	}

	return obj.SetCacheObject(cacheData)
}

func (r *RedisCacheProvider) SaveObj(obj CacheableObject) error {
	return r.Save(obj.CacheKey(), obj, obj.CacheTTL())
}

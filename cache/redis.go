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
	"strings"
	"time"
)

// InfiniteTTL - do th legacy, it needs to a huge positive number
const InfiniteTTL = 99999 * time.Hour

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
	if strings.Index(key, r.prefix) == 0 {
		return key
	}

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

	return nil
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

func (r *RedisCacheProvider) saveMap(key string, val interface{}, ttl time.Duration) error {
	err := r.Client().HMSet(r.ctx, r.toKey(key), val).Err()
	if err != nil {
		return err
	}

	_, err = r.UpdateTTl(key, ttl)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisCacheProvider) Save(key string, val any, ttl time.Duration) error {
	// check if val is a map
	switch valT := val.(type) {
	case string:
		val = valT
	case []byte:
		val = string(valT)
	case int, int32, int64, uint, uint32, uint64:
		val = fmt.Sprintf("%d", val)
	case float32, float64:
		val = fmt.Sprintf("%f", val)
	case CacheableObject:
		return r.saveMap(key, valT.GetCacheObject(), ttl)
	default:
		if reflect.TypeOf(val).Kind() == reflect.Map {
			return r.saveMap(key, val, ttl)
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			res, err := r.Client().SAdd(r.ctx, r.toKey(key), val.([]any)...).Result()
			if err != nil {
				return err
			}

			if res == 0 && len(val.([]any)) > 0 {
				return errors.New("failed to save set")
			}
		} else {
			return fmt.Errorf("unsupported type %T", val)
		}
	}

	return r.Client().Set(r.ctx, r.toKey(key), val, ttl).Err()
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

	return r.Save(key, valJson, ttl)
}

func (r *RedisCacheProvider) LoadJSON(key string, result any) (any, error) {
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

func (r *RedisCacheProvider) SaveMap(key string, val map[string]string, ttl time.Duration) error {
	return r.saveMap(key, val, ttl)
}

func (r *RedisCacheProvider) GetMap(key string) (map[string]string, error) {
	return r.Client().HGetAll(r.ctx, r.toKey(key)).Result()
}

func (r *RedisCacheProvider) GetMapKeys(key string, fields ...string) (map[string]string, error) {
	resultMap := map[string]string{}
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

func (r *RedisCacheProvider) SaveSet(key string, members []string, ttl time.Duration) error {
	_, err := r.Client().SAdd(r.ctx, r.toKey(key), members).Result()
	if err != nil {
		return err
	}

	_, err = r.UpdateTTl(key, ttl)
	if err != nil {
		return err
	}

	return nil
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

func (r *RedisCacheProvider) InSet(key string, member string) (bool, error) {
	return r.Client().SIsMember(r.ctx, r.toKey(key), member).Result()
}

func (r *RedisCacheProvider) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *RedisCacheProvider) GetPrefix() string {
	return r.prefix
}

func (r *RedisCacheProvider) LoadObj(obj CacheableObject) error {
	key := obj.CacheKey()
	if r.MustExists(key) == false {
		return KeyNotFound{Key: key}
	}

	cacheData, err := r.GetMap(key)
	if err != nil {
		return err
	}

	return obj.SetCacheObject(cacheData)
}

func (r *RedisCacheProvider) SaveObj(obj CacheableObject) error {
	return r.saveMap(obj.CacheKey(), obj.GetCacheObject(), obj.CacheTTL())
}

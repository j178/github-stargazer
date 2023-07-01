package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/redis/rueidis"

	"github.com/j178/github_stargazer/backend/config"
)

const (
	RedisKeyPrefix    = "github_stargazer"
	DefaultExpiration = 10 * time.Minute
)

const (
	DEFAULT = time.Duration(0)
	FOREVER = time.Duration(-1)
)

// CacheStore is the interface of a Default backend
type CacheStore interface {
	// Get retrieves an item from the Default. Returns the item or nil, and a bool indicating
	// whether the key was found.
	Get(ctx context.Context, key string, value interface{}) error

	// Set sets an item to the Default, replacing any existing item.
	Set(ctx context.Context, key string, value interface{}, expires time.Duration) error

	// Delete removes an item from the Default. Does nothing if the key is not in the Default.
	Delete(ctx context.Context, key string) error
}

type redisCache struct {
	redis             rueidis.Client
	defaultExpiration time.Duration
}

func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	cmd := c.redis.B().Get().Key(key).Build()
	v, err := c.redis.Do(ctx, cmd).AsBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(v, value)
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, expires time.Duration) error {
	if expires == DEFAULT {
		expires = c.defaultExpiration
	}

	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var cmd rueidis.Completed
	if expires > 0 {
		cmd = c.redis.B().Set().Key(key).Value(string(v)).Ex(expires).Build()
	} else {
		cmd = c.redis.B().Set().Key(key).Value(string(v)).Build()
	}
	err = c.redis.Do(ctx, cmd).Error()
	return err
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	cmd := c.redis.B().Del().Key(key).Build()
	err := c.redis.Do(ctx, cmd).Error()
	return err
}

func getRedis() (rueidis.Client, error) {
	u, err := url.Parse(config.KvURL)
	if err != nil {
		log.Fatalf("parse kv url failed: %s", err)
	}
	username := u.User.Username()
	passwd, _ := u.User.Password()
	host, _, _ := net.SplitHostPort(u.Host)
	opt := rueidis.ClientOption{
		ForceSingleClient: true,
		DisableCache:      true,
		Username:          username,
		Password:          passwd,
		InitAddress: []string{
			u.Host,
		},
		TLSConfig: &tls.Config{
			ServerName: host,
		},
	}

	r, err := rueidis.NewClient(opt)
	return r, err
}

func keyFunc(name string, id string) string {
	return fmt.Sprintf("%s:%s:%s", RedisKeyPrefix, name, id)
}

var defaultCache CacheStore
var cacheOnce sync.Once

func Default() CacheStore {
	cacheOnce.Do(
		func() {
			defaultCache = &redisCache{
				redis:             Redis(),
				defaultExpiration: DefaultExpiration,
			}
		},
	)
	return defaultCache
}

var redis rueidis.Client
var redisOnce sync.Once

func Redis() rueidis.Client {
	redisOnce.Do(
		func() {
			r, err := getRedis()
			if err != nil {
				log.Fatalf("create redis client failed: %s", err)
			}
			redis = r
		},
	)
	return redis
}

func Get(ctx context.Context, name, key string, value interface{}) error {
	return Default().Get(ctx, keyFunc(name, key), value)
}

func Set(ctx context.Context, name, key string, value interface{}, expire time.Duration) error {
	return Default().Set(ctx, keyFunc(name, key), value, expire)
}

func Delete(ctx context.Context, name, key string) error {
	return Default().Delete(ctx, keyFunc(name, key))
}

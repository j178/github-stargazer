package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
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

var (
	ErrCacheMiss = fmt.Errorf("cache: key not found")
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

	Incr(ctx context.Context, key string, value *int64, expires time.Duration) error
}

type redisCache struct {
	redis             rueidis.Client
	defaultExpiration time.Duration
}

func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	cmd := c.redis.B().Get().Key(key).Build()
	v, err := c.redis.Do(ctx, cmd).AsBytes()
	if rueidis.IsRedisNil(err) {
		return ErrCacheMiss
	}
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

func (c *redisCache) Incr(ctx context.Context, key string, value *int64, expires time.Duration) error {
	if expires == DEFAULT {
		expires = c.defaultExpiration
	}

	var err error
	if expires > 0 {
		cmds := []rueidis.Completed{
			c.redis.B().Incr().Key(key).Build(),
			c.redis.B().Expire().Key(key).Seconds(int64(expires / time.Second)).Build(),
		}
		vals := c.redis.DoMulti(ctx, cmds...)
		for _, v := range vals {
			if v.Error() != nil {
				return v.Error()
			}
		}
		*value, err = vals[0].AsInt64()
		return err
	} else {
		cmd := c.redis.B().Incr().Key(key).Build()
		*value, err = c.redis.Do(ctx, cmd).AsInt64()
		return err
	}
}

func initRedis() (rueidis.Client, error) {
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

type Key []string

func (k Key) String() string {
	if len(k) < 2 {
		panic("key must be at least 2 parts")
	}
	return fmt.Sprintf("%s:%s", RedisKeyPrefix, strings.Join(k, ":"))
}

var Default = sync.OnceValue(
	func() CacheStore {
		return &redisCache{
			redis:             Redis(),
			defaultExpiration: DefaultExpiration,
		}
	},
)

var Redis = sync.OnceValue(
	func() rueidis.Client {
		r, err := initRedis()
		if err != nil {
			log.Fatalf("create redis client: %s", err)
		}
		return r
	},
)

func Get[T any](ctx context.Context, key Key) (T, error) {
	var z T
	err := Default().Get(ctx, key.String(), &z)
	return z, err
}

func Set[T any](ctx context.Context, key Key, value T, expire time.Duration) error {
	return Default().Set(ctx, key.String(), value, expire)
}

func Delete(ctx context.Context, key Key) error {
	return Default().Delete(ctx, key.String())
}

func Incr(ctx context.Context, key Key, expire time.Duration) (int64, error) {
	var z int64
	err := Default().Incr(ctx, key.String(), &z, expire)
	return z, err
}

func GetOrCreate[T any](
	ctx context.Context,
	key Key,
	expire time.Duration,
	create func() (T, error),
) (T, error) {
	var z T
	value, err := Get[T](ctx, key)
	if err == nil {
		return value, nil
	}
	if err != ErrCacheMiss {
		return z, err
	}

	v, err := create()
	if err != nil {
		return z, err
	}

	err = Set(ctx, key, v, expire)

	return v, err
}

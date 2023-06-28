package api

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"

	"github.com/redis/rueidis"
)

func envOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func env(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("environment variable %s is not set", key))
}

var (
	webhookSecret = envOrDefault("GITHUB_WEBHOOK_SECRET", "")
	clientID      = env("GITHUB_CLIENT_ID")
	clientSecret  = env("GITHUB_CLIENT_SECRET")
	kvURL         = env("KV_URL")
)

var redis rueidis.Client

func init() {
	u, err := url.Parse(kvURL)
	if err != nil {
		panic(err)
	}
	passwd, _ := u.User.Password()
	opt := rueidis.ClientOption{
		Username: u.User.Username(),
		Password: passwd,
		InitAddress: []string{
			u.Host,
		},
	}
	if u.Scheme == "rediss" {
		opt.TLSConfig = &tls.Config{
			ServerName: u.Host,
		}
	}
	redis, err = rueidis.NewClient(opt)
	if err != nil {
		panic(err)
	}
}

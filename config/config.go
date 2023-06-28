package config

import (
	"fmt"
	"os"
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
	WebhookSecret = envOrDefault("GITHUB_WEBHOOK_SECRET", "")
	ClientID      = env("GITHUB_CLIENT_ID")
	ClientSecret  = env("GITHUB_CLIENT_SECRET")
	KvURL         = env("KV_URL")
)

package config

import (
	"log"
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
	log.Fatalf("environment variable %s is not set", key)
	return ""
}

var (
	WebhookSecret = envOrDefault("GITHUB_WEBHOOK_SECRET", "")
	ClientID      = env("GITHUB_CLIENT_ID")
	ClientSecret  = env("GITHUB_CLIENT_SECRET")
	KvURL         = env("KV_URL")
	SecretKey     = env("SECRET_KEY")
)

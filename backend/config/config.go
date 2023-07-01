package config

import (
	"log"
	"os"
	"strconv"
	"sync"
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
	AppID            int64
	AppPrivateKey    []byte
	ClientID         string
	ClientSecret     string
	WebhookSecret    []byte
	KvURL            string
	SecretKey        []byte
	TelegramBotToken string
	VercelURL        string
)

func loadEnv() {
	appIdStr := env("GITHUB_APP_ID")
	var err error
	AppID, err = strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		log.Fatalf("parse GITHUB_APP_ID: %v", err)
	}

	AppPrivateKey = []byte(env("GITHUB_APP_PRIVATE_KEY"))
	ClientID = env("GITHUB_CLIENT_ID")
	ClientSecret = env("GITHUB_CLIENT_SECRET")
	WebhookSecret = []byte(envOrDefault("GITHUB_WEBHOOK_SECRET", ""))
	KvURL = env("KV_URL")
	SecretKey = []byte(env("SECRET_KEY"))
	TelegramBotToken = env("TELEGRAM_BOT_TOKEN")
	VercelURL = "https://" + env("VERCEL_URL")
}

var once sync.Once

func Load() {
	once.Do(loadEnv)
}

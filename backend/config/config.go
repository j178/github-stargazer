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

const (
	defaultTelegramBotUsername = "gh_stargazer_bot"
)

var (
	BaseURL             string
	AppID               int64
	AppPrivateKey       []byte
	ClientID            string
	ClientSecret        string
	WebhookSecret       []byte
	KvURL               string
	SecretKey           []byte
	TelegramBotToken    string
	TelegramBotUsername string
	DiscordBotToken     string
)

func loadEnv() {
	appIdStr := env("GITHUB_APP_ID")
	var err error
	AppID, err = strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		log.Fatalf("parse GITHUB_APP_ID: %v", err)
	}

	BaseURL = env("BASE_URL")
	AppPrivateKey = []byte(env("GITHUB_APP_PRIVATE_KEY"))
	ClientID = env("GITHUB_CLIENT_ID")
	ClientSecret = env("GITHUB_CLIENT_SECRET")
	WebhookSecret = []byte(envOrDefault("GITHUB_WEBHOOK_SECRET", ""))
	KvURL = env("KV_URL")
	SecretKey = []byte(env("SECRET_KEY"))
	TelegramBotToken = env("TELEGRAM_BOT_TOKEN")
	TelegramBotUsername = envOrDefault("TELEGRAM_BOT_USERNAME", defaultTelegramBotUsername)
	DiscordBotToken = env("DISCORD_BOT_TOKEN")
}

var once sync.Once

func Load() {
	once.Do(loadEnv)
}

package notify

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/bark"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/telegram"
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

func Notify(ctx context.Context, subject, message string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("notify: %v", e)
		}
	}()

	notifier := notify.New()
	service := envOrDefault("NOTIFY_SERVICE", "bark")
	switch service {
	case "bark":
		notifier.UseServices(bark.NewWithServers(env("BARK_KEY"), envOrDefault("BARK_SERVER", bark.DefaultServerURL)))
	case "telegram":
		tg, err := telegram.New(env("TELEGRAM_TOKEN"))
		if err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
		tg.SetParseMode("MarkdownV2")
		chatIDStr := env("TELEGRAM_CHAT_ID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("telegram parse chat_id: %w", err)
		}
		tg.AddReceivers(chatID)
		notifier.UseServices(tg)
	case "discord":
		ds := discord.New()
		ds.AddReceivers(env("DISCORD_CHANNEL_ID"))
		notifier.UseServices(ds)
	}
	err = notifier.Send(ctx, subject, message)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

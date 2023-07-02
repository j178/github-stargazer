package notify

import (
	"fmt"

	"github.com/nikoksr/notify"
)

// 参考 github.com/nikoksr/notify, github.com/goreleaser/goreleaser 和 github.com/megaease/easeprobe

type Notifier interface {
	notify.Notifier
	Configure(settings map[string]string) error
}

func GetNotifier(settings []map[string]string) (*notify.Notify, error) {
	notifier := notify.New()
	for _, setting := range settings {
		service := setting["service"]
		// TODO: add slack, mastodon, etc.
		switch service {
		case "bark":
			bark := &barkService{}
			err := bark.Configure(setting)
			if err != nil {
				return nil, fmt.Errorf("bark: %w", err)
			}
			notifier.UseServices(bark)
		case "telegram":
			tg := &telegramService{}
			err := tg.Configure(setting)
			if err != nil {
				return nil, fmt.Errorf("telegram: %w", err)
			}
			notifier.UseServices(tg)
		case "discord":
			discord := &discordService{}
			err := discord.Configure(setting)
			if err != nil {
				return nil, fmt.Errorf("discord: %w", err)
			}
			notifier.UseServices(discord)
		case "webhook":
			webhook := &webhookService{}
			err := webhook.Configure(setting)
			if err != nil {
				return nil, fmt.Errorf("webhook: %w", err)
			}
			notifier.UseServices(webhook)
		default:
			return nil, fmt.Errorf("unknown service: %s", service)
		}
	}

	return notifier, nil
}

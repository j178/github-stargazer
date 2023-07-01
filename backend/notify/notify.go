package notify

import (
	"fmt"

	"github.com/nikoksr/notify"
)

type Notifier interface {
	notify.Notifier
	FromSettings(settings map[string]string) error
}

func GetNotifier(settings []map[string]string) (*notify.Notify, error) {
	notifier := notify.New()
	for _, setting := range settings {
		service := setting["service"]
		switch service {
		case "bark":
			bark := &barkService{}
			err := bark.FromSettings(setting)
			if err != nil {
				return nil, fmt.Errorf("bark: %w", err)
			}
			notifier.UseServices(bark)
		case "telegram":
			tg := &telegramService{}
			err := tg.FromSettings(setting)
			if err != nil {
				return nil, fmt.Errorf("telegram: %w", err)
			}
			notifier.UseServices(tg)
		case "discord":
			discord := &discordService{}
			err := discord.FromSettings(setting)
			if err != nil {
				return nil, fmt.Errorf("discord: %w", err)
			}
			notifier.UseServices(discord)
		case "http":
			http := &httpService{}
			err := http.FromSettings(setting)
			if err != nil {
				return nil, fmt.Errorf("http: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown service: %s", service)
		}
	}

	return notifier, nil
}

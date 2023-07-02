package notify

import (
	"fmt"

	"github.com/nikoksr/notify"
)

// 参考 github.com/nikoksr/notify, github.com/goreleaser/goreleaser 和 github.com/megaease/easeprobe

type Notifier interface {
	notify.Notifier
	Name() string
	Configure(settings map[string]string) error
}

func GetNotifier(settings []map[string]string) (*notify.Notify, error) {
	notifier := notify.New()
	for _, setting := range settings {
		serviceName := setting["service"]
		var service Notifier
		// TODO: add slack, mastodon, etc.
		switch serviceName {
		case "bark":
			service = &barkService{}
		case "telegram":
			service = &telegramService{}
		case "discord":
			service = &discordService{}
		case "webhook":
			service = &webhookService{}
		default:
			return nil, fmt.Errorf("unknown service: %s", service)
		}

		err := service.Configure(setting)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", service.Name(), err)
		}
		notifier.UseServices(service)
	}

	return notifier, nil
}

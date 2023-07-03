package notify

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

// 参考 github.com/nikoksr/notify, github.com/goreleaser/goreleaser 和 github.com/megaease/easeprobe

type Notifier interface {
	Name() string
	Configure(settings map[string]string) error
	Send(context.Context, string, string) error
}

type Notify struct {
	notifiers []Notifier
}

func (n *Notify) AddNotifier(notifier Notifier) {
	n.notifiers = append(n.notifiers, notifier)
}

func (n *Notify) Send(ctx context.Context, subject, message string) error {
	wg := pool.New().WithContext(ctx).WithMaxGoroutines(10)
	for _, notifier := range n.notifiers {
		if notifier == nil {
			continue
		}

		notifier := notifier
		wg.Go(
			func(ctx context.Context) error {
				return notifier.Send(ctx, subject, message)
			},
		)
	}

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("send notify: %w", err)
	}

	return nil
}

func GetNotifier(settings []map[string]string) (*Notify, error) {
	notify := &Notify{}
	for _, setting := range settings {
		serviceName := setting["service"]
		var service Notifier
		// TODO: add slack, mastodon, etc.
		switch serviceName {
		case "bark":
			service = &barkService{}
		case "telegram":
			service = &telegramService{}
		case "discord_webhook":
			service = &discordWebhookService{}
		case "discord_bot":
			service = &discordBotService{}
		case "webhook":
			service = &webhookService{}
		default:
			return nil, fmt.Errorf("unknown service: %s", service)
		}

		err := service.Configure(setting)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", service.Name(), err)
		}
		notify.AddNotifier(service)
	}

	return notify, nil
}

package notify

import (
	"fmt"
	"strconv"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/bark"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/telegram"
)

func GetNotifier(settings []map[string]string) (*notify.Notify, error) {
	notifier := notify.New()
	for _, setting := range settings {
		service := setting["service"]
		switch service {
		case "bark":
			key := setting["key"]
			server := setting["server"]
			if key == "" {
				return nil, fmt.Errorf("bark: key is empty")
			}
			if server == "" {
				server = bark.DefaultServerURL
			}
			barkService := bark.NewWithServers(key, server)
			notifier.UseServices(barkService)
		case "telegram":
			token := setting["token"]
			chatIDStr := setting["chat_id"]
			if token == "" || chatIDStr == "" {
				return nil, fmt.Errorf("telegram: token or chat_id is empty")
			}
			chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("telegram: parse chat_id: %w", err)
			}
			telegramService, err := telegram.New(token)
			if err != nil {
				return nil, fmt.Errorf("telegram: %w", err)
			}
			telegramService.AddReceivers(chatID)
			notifier.UseServices(telegramService)
		case "discord":
			// TODO support discord webhook
			webhookURL := setting["webhook_url"]
			if webhookURL == "" {
				return nil, fmt.Errorf("discord: webhook_url is empty")
			}
			discordService := discord.New()
			discordService.AddReceivers(webhookURL)
		default:
			return nil, fmt.Errorf("unknown service: %s", service)
		}
	}

	return notifier, nil
}

package notify

import (
	"context"
	"log"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/pkg/errors"
)

var defaultTgBotOnce sync.Once
var defaultTgBot *tgbotapi.BotAPI

// init bot will make a `getMe` request, so we cache it

func DefaultTelegramBot() *tgbotapi.BotAPI {
	defaultTgBotOnce.Do(
		func() {
			var err error
			defaultTgBot, err = tgbotapi.NewBotAPI(config.TelegramBotToken)
			if err != nil {
				log.Fatal("init telegram bot: %w", err)
			}
		},
	)
	return defaultTgBot
}

type telegramService struct {
	client *tgbotapi.BotAPI
	chatID int64
}

func (t *telegramService) Name() string {
	return "telegram"
}

func (t *telegramService) Configure(settings map[string]string) error {
	token := settings["token"]
	chatIDStr := settings["chat_id"]
	if chatIDStr == "" {
		return errors.New("chat_id is empty")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return errors.New("invalid chat_id")
	}

	var tg *tgbotapi.BotAPI
	if token == "" || token == "default" {
		tg = DefaultTelegramBot()
	} else {
		tg, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			return err
		}
	}

	t.client = tg
	t.chatID = chatID

	return nil
}

func (t *telegramService) Send(ctx context.Context, subject, message string) error {
	fullMessage := subject + "\n" + message // Treating subject as message title

	msg := tgbotapi.NewMessage(t.chatID, fullMessage)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = true

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := t.client.Send(msg)
		if err != nil {
			return errors.Wrapf(err, "telegram: failed to send message to Telegram chat '%d'", t.chatID)
		}
	}

	return nil
}

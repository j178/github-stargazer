package notify

import (
	"context"
	"fmt"
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
	client  *tgbotapi.BotAPI
	chatIDs []int64
}

func (t *telegramService) FromSettings(settings map[string]string) error {
	token := settings["token"]
	chatIDStr := settings["chat_id"]
	if chatIDStr == "" {
		return errors.New("telegram: chat_id is empty")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return errors.New("telegram: invalid chat_id")
	}

	var tg *tgbotapi.BotAPI
	if token == "" || token == "default" {
		tg = DefaultTelegramBot()
	} else {
		tg, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
	}

	t.client = tg
	t.AddReceivers(chatID)

	return nil
}

// AddReceivers takes Telegram chat IDs and adds them to the internal chat ID list. The Send method will send
// a given message to all those chats.
func (t *telegramService) AddReceivers(chatIDs ...int64) {
	t.chatIDs = append(t.chatIDs, chatIDs...)
}

// Send takes a message subject and a message body and sends them to all previously set chats. Message body supports
// html as markup language.
func (t *telegramService) Send(ctx context.Context, subject, message string) error {
	fullMessage := subject + "\n" + message // Treating subject as message title

	msg := tgbotapi.NewMessage(0, fullMessage)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = true

	for _, chatID := range t.chatIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg.ChatID = chatID
			_, err := t.client.Send(msg)
			if err != nil {
				return errors.Wrapf(err, "failed to send message to Telegram chat '%d'", chatID)
			}
		}
	}

	return nil
}

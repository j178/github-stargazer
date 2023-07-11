package telegram

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/j178/github_stargazer/backend/routes"
	"github.com/j178/github_stargazer/backend/routes/configure"

	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/notify"
)

func setupWebhook() {
	bot := notify.DefaultTelegramBot()
	webhook, _ := tgbotapi.NewWebhook(config.BaseURL + "/api/webhook/telegram")

	_, err := bot.Request(webhook)
	if err != nil {
		log.Fatal("set telegram webhook: %w", err)
	}
}

var once sync.Once

func Bot() *tgbotapi.BotAPI {
	once.Do(setupWebhook)
	return notify.DefaultTelegramBot()
}

func OnUpdate(c *gin.Context) {
	update, err := Bot().HandleUpdate(c.Request)
	if err != nil {
		routes.Abort(c, http.StatusBadRequest, err, "handle update")
		return
	}

	log.Printf("update: %+v", update)

	var message *tgbotapi.Message
	if update.Message != nil {
		message = update.Message
	} else if update.EditedMessage != nil {
		message = update.EditedMessage
	} else if update.ChannelPost != nil {
		message = update.ChannelPost
	} else if update.EditedChannelPost != nil {
		message = update.EditedChannelPost
	}

	if message == nil || !strings.HasPrefix(message.Text, "/start") {
		c.JSON(http.StatusOK, gin.H{"status": "not /start command"})
		return
	}

	chatID := message.Chat.ID
	tgUsername := message.From.UserName
	replyTo := message.MessageID
	commands := strings.SplitN(strings.TrimSpace(message.Text), " ", 2)
	startArgs := ""
	if len(commands) > 1 {
		startArgs = commands[1]
	}

	if startArgs == "" {
		msg := "Send `/start <connect token>` to connect your GitHub account"
		if message.Chat.IsGroup() || message.Chat.IsSuperGroup() {
			msg = fmt.Sprintf(
				"Send `/start@%s <connect token>` to connect your GitHub account",
				config.TelegramBotUsername,
			)
		}
		reply := tgbotapi.NewMessage(chatID, msg)
		reply.ReplyToMessageID = replyTo
		reply.ParseMode = "MarkdownV2"

		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	connect := map[string]any{
		"chat_id":           chatID,
		"telegram_username": tgUsername,
	}

	err = configure.SetConnectResult(c, startArgs, "telegram", connect)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid connect token"})

		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Invalid connect token: %q", startArgs))
		reply.ReplyToMessageID = replyTo

		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	reply := tgbotapi.NewMessage(chatID, "Connected!")
	reply.ReplyToMessageID = replyTo

	_, err = Bot().Send(reply)
	if err != nil {
		log.Printf("send message: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

package telegram

import (
	"fmt"
	"log"
	"net/http"
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
	webhook, _ := tgbotapi.NewWebhook(config.BaseURL + "/api/telegram")

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

	if update.Message.Command() != "start" {
		c.JSON(http.StatusOK, gin.H{"status": "not /start command"})
		return
	}

	chatID := update.Message.Chat.ID
	tgUsername := update.Message.From.UserName
	replyTo := update.Message.MessageID
	commandArgs := update.Message.CommandArguments()

	if commandArgs == "" {
		msg := "Send `/start <connect token>` to connect your GitHub account"
		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			msg = fmt.Sprintf(
				"Send `/start@%s <connect token>` to connect your GitHub account",
				config.TelegramBotUsername,
			)
		}
		reply := tgbotapi.NewMessage(chatID, msg)
		reply.ReplyToMessageID = replyTo

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

	err = configure.SetConnectResult(c, commandArgs, connect)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid connect token"})

		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Invalid connect token: %q", commandArgs))
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

package routes

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/j178/github_stargazer/backend/cache"
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

const (
	Issuer      = "telegram-connect"
	BotUsername = "gh_stargazer_bot"
)

func GenerateTelegramConnectToken(c *gin.Context) {
	login := c.GetString("login")

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   login,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
		},
	)
	tokenString, _ := token.SignedString(config.SecretKey)
	c.JSON(
		http.StatusOK, gin.H{
			"bot_url":        "https://t.me/" + BotUsername,
			"connect_string": tokenString,
		},
	)
}

// 轮询获取绑定结果

func GetTelegramConnect(c *gin.Context) {
	login := c.GetString("login")

	connect, err := cache.Get[map[string]any](c, cache.Key{"telegram_connect", login})
	if err == cache.ErrCacheMiss {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "get cache")
		return
	}

	c.JSON(http.StatusOK, connect)
}

func selfJoinChat(update *tgbotapi.Update) bool {
	if update.Message == nil {
		return false
	}

	for _, member := range update.Message.NewChatMembers {
		if member.UserName == BotUsername {
			return true
		}
	}

	return false
}

func OnTelegramUpdate(c *gin.Context) {
	update, err := Bot().HandleUpdate(c.Request)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "handle update")
		return
	}

	if update.Message == nil || (update.Message.Text == "" && !selfJoinChat(update)) {
		c.JSON(http.StatusOK, gin.H{"status": "not a message"})
		return
	}

	chatID := update.FromChat().ID
	tgUsername := update.SentFrom().UserName
	replyTo := update.Message.MessageID
	text := strings.TrimSpace(update.Message.Text)

	// strip command prefix
	if update.Message.IsCommand() {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 {
			text = strings.TrimSpace(parts[1])
		} else {
			text = ""
		}
	}

	if text == "" {
		msg := "Hi, welcome to use GitHub Stargazer Bot. Please send your connect string."
		if update.Message.Chat.IsGroup() {
			msg = "Hi, welcome to use GitHub Stargazer Bot. Please reply this message with your connect string."
		}
		reply := tgbotapi.NewMessage(chatID, msg)
		reply.ReplyToMessageID = replyTo

		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	v, err := jwt.ParseWithClaims(
		text, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
			return config.SecretKey, nil
		},
		jwt.WithIssuer(Issuer),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		log.Printf("parse connect string: %v", err)
		c.JSON(http.StatusOK, gin.H{"error": "invalid connect string"})

		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Invalid connect string: %q", text))
		reply.ReplyToMessageID = replyTo

		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	claims, _ := v.Claims.(*jwt.RegisteredClaims)
	account := claims.Subject

	connect := map[string]any{
		"account":           account,
		"chatID":            chatID,
		"telegram_username": tgUsername,
	}
	err = cache.Set(c, cache.Key{"telegram_connect", account}, connect, 10*time.Minute)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "set cache")

		reply := tgbotapi.NewMessage(chatID, "internal server error")
		reply.ReplyToMessageID = replyTo

		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	reply := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("Bind to %s, new star notifications will be sent here.", account),
	)
	reply.ReplyToMessageID = replyTo

	_, err = Bot().Send(reply)
	if err != nil {
		log.Printf("send message: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

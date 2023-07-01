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
)

func setup() {
	var err error
	bot, err = tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatal("init telegram bot: %w", err)
	}

	webhook, _ := tgbotapi.NewWebhook(config.BaseURL + "/api/telegram")

	_, err = bot.Request(webhook)
	if err != nil {
		log.Fatal("set telegram webhook: %w", err)
	}
}

var bot *tgbotapi.BotAPI
var once sync.Once

func Bot() *tgbotapi.BotAPI {
	once.Do(setup)
	return bot
}

const (
	Issuer = "telegram-connect"
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
	c.JSON(http.StatusOK, gin.H{"connect_string": tokenString})
}

// 轮询获取绑定结果

func GetTelegramConnect(c *gin.Context) {
	login := c.GetString("login")

	var connect map[string]any
	err := cache.Get(c, "telegram_connect", login, &connect)
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

func OnTelegramUpdate(c *gin.Context) {
	update, err := Bot().HandleUpdate(c.Request)
	if err != nil {
		Abort(c, http.StatusBadRequest, err, "handle update")
		return
	}

	if update.Message == nil {
		c.JSON(http.StatusOK, gin.H{"status": "not a message"})
		return
	}

	chatID := update.Message.Chat.ID
	tgUsername := update.Message.From.UserName

	text := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/start"))
	v, err := jwt.ParseWithClaims(
		text, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
			return config.SecretKey, nil
		},
		jwt.WithIssuer(Issuer),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "not connect string"})

		reply := tgbotapi.NewMessage(chatID, "this is not a connect string")
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
	err = cache.Set(c, "telegram_connect", account, connect, 10*time.Minute)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "set cache")

		reply := tgbotapi.NewMessage(chatID, "internal server error")
		_, err = Bot().Send(reply)
		if err != nil {
			log.Printf("send message: %v", err)
		}
		return
	}

	reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("bind to %s, chat_id=%d", account, chatID))
	_, err = Bot().Send(reply)
	if err != nil {
		log.Printf("send message: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

package configure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/j178/github_stargazer/backend/cache"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/routes"
	"github.com/j178/github_stargazer/backend/utils"
)

const ConnectTokenExpire = 10 * time.Minute

func GenerateConnectToken(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		routes.Abort(c, http.StatusBadRequest, nil, "platform is empty")
		return
	}

	token, err := utils.GenerateRandomString(32)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "generate token")
		return
	}
	err = cache.Set(c, cache.Key{"connect", token}, map[string]any{}, ConnectTokenExpire)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "set cache")
		return
	}

	resp := gin.H{
		"token":  token,
		"expire": time.Now().Add(ConnectTokenExpire).Unix(),
	}

	switch platform {
	case "telegram":
		resp["bot_url"] = fmt.Sprintf("https://t.me/%s?start=%s", config.TelegramBotUsername, token)
		resp["bot_group_url"] = fmt.Sprintf("https://t.me/%s?startgroup=%s", config.TelegramBotUsername, token)
	case "discord":
		// TODO
	case "slack":
		// TODO
	}
	c.JSON(http.StatusOK, resp)
}

func GetConnectResult(c *gin.Context) {
	platform := c.Param("platform")
	token := c.Param("token")
	if platform == "" || token == "" {
		routes.Abort(c, http.StatusBadRequest, nil, "platform or token is empty")
		return
	}

	r, err := cache.Get[map[string]any](c, cache.Key{"connect", token})
	if err == cache.ErrCacheMiss {
		routes.Abort(c, http.StatusNotFound, err, "token not found")
		return
	}
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "get cache")
		return
	}

	c.JSON(http.StatusOK, r)
}

func SetConnectResult(ctx context.Context, token string, result map[string]any) error {
	_, err := cache.Get[map[string]any](ctx, cache.Key{"connect", token})
	if err != nil {
		return err
	}

	err = cache.Set(ctx, cache.Key{"connect", token}, result, ConnectTokenExpire)
	if err != nil {
		return err
	}

	return err
}
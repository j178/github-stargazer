package configure

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/j178/github_stargazer/backend/cache"
	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/routes"
	"github.com/j178/github_stargazer/backend/utils"
)

const ConnectTokenExpire = 10 * time.Minute

var ConnectTokenPlatforms = []string{"telegram", "discord", "slack"}

func GenerateConnectToken(c *gin.Context) {
	login := c.GetString("login")
	platform := c.Param("platform")
	if !slices.Contains(ConnectTokenPlatforms, platform) {
		routes.Abort(c, http.StatusBadRequest, nil, "invalid platform")
		return
	}

	count, err := cache.Get[int64](c, cache.Key{"connect_token_count", login, platform})
	if err != nil && err != cache.ErrCacheMiss {
		routes.Abort(c, http.StatusInternalServerError, err, "get connect token count")
		return
	}
	if count >= 2 {
		routes.Abort(c, http.StatusForbidden, nil, "exceed connect token count limit")
		return
	}

	token, err := utils.GenerateRandomString(32)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "generate token")
		return
	}
	err = cache.Set(c, cache.Key{"connect", token}, map[string]any{"platform": platform}, ConnectTokenExpire)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "set cache")
		return
	}

	_, err = cache.Incr(c, cache.Key{"connect_token_count", login, platform}, 60*time.Second)
	if err != nil {
		routes.Abort(c, http.StatusInternalServerError, err, "incr connect token count")
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
		resp["bot_url"] = fmt.Sprintf(
			"https://discord.com/api/oauth2/authorize?client_id=%s&permissions=2048&scope=bot%%20applications.commands",
			config.DiscordAppID,
		)
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

func SetConnectResult(ctx context.Context, token string, platform string, result map[string]any) error {
	prev, err := cache.Get[map[string]any](ctx, cache.Key{"connect", token})
	if err != nil {
		return err
	}
	if len(prev) > 1 { // not only `platform`
		return fmt.Errorf("token already used")
	}
	if p := prev["platform"]; p == nil || p.(string) != platform {
		return fmt.Errorf("platform not match")
	}

	err = cache.Set(ctx, cache.Key{"connect", token}, result, ConnectTokenExpire)
	if err != nil {
		return err
	}

	return err
}

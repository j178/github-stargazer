package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/config"
)

// 开启 "Request user authorization (OAuth) during installation" 之后，安装的过程同时也是授权的过程
// 用户授权之后，GitHub 会带 code 将用户重定向到这里
// 这里设置 cookie session 后重定向回到之前的页面

func Authorized(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "code is empty")
		return
	}
	state := c.Query("state")
	if state == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "state is empty")
		return
	}
	returnUrl, err := decodeState(state, config.SecretKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     oauthGitHub.Endpoint,
	}
	token, err := cfg.Exchange(c, code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// load user info
	client := github.NewTokenClient(context.Background(), token.AccessToken)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"access_token": token.AccessToken,
			"login":        user.Login,
		},
	)
	session, err := jwtToken.SignedString([]byte(config.SecretKey))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("session", session, 90000, "/", "", true, true)
	c.Redirect(http.StatusFound, returnUrl)
}

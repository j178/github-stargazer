package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v53/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/backend/cache"
	"github.com/j178/github_stargazer/backend/config"
)

// 开启 "Request user authorization (OAuth) during installation" 之后，安装的过程同时也是授权的过程
// 用户授权之后，GitHub 会带 code 将用户重定向到这里
// 这里设置 cookie session 后重定向回到之前的页面

func Abort(c *gin.Context, code int, err error, msg string) {
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if err == nil {
		err = errors.New(msg)
	} else {
		err = errors.WithMessage(err, msg)
	}
	c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
}

func Authorized(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		Abort(c, http.StatusBadRequest, nil, "code is empty")
		return
	}
	returnUrl := "/"
	state := c.Query("state")
	// Install & Authorize redirects do not include `state`
	if state != "" {
		var err error
		returnUrl, err = decodeState(state, config.SecretKey)
		if err != nil {
			Abort(c, http.StatusUnauthorized, err, "decode state")
			return
		}
	}

	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     oauthGitHub.Endpoint,
	}
	token, err := cfg.Exchange(c, code)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "exchange token")
		return
	}

	// load user info
	client := github.NewTokenClient(c, token.AccessToken)
	user, _, err := client.Users.Get(c, "")
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "get user info")
		return
	}

	err = cache.SaveOAuthToken(c, user.GetLogin(), token)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "save oauth token")
		return
	}

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   user.GetLogin(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	)
	session, err := jwtToken.SignedString(config.SecretKey)
	if err != nil {
		Abort(c, http.StatusInternalServerError, err, "generate session")
		return
	}

	// refresh token expires in 6 months
	c.SetCookie("session", session, 6*30*24*3600, "/", "", true, true)
	c.Redirect(http.StatusFound, returnUrl)
}

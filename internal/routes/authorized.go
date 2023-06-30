package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/internal/middleware"

	"github.com/j178/github_stargazer/internal/config"
)

// 开启 "Request user authorization (OAuth) during installation" 之后，安装的过程同时也是授权的过程
// 用户授权之后，GitHub 会带 code 将用户重定向到这里
// 这里设置 cookie session 后重定向回到之前的页面

func Authorized(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "code is empty"})
		return
	}
	returnUrl := "/"
	state := c.Query("state")
	// Install & Authorize redirects do not include `state`
	if state != "" {
		var err error
		returnUrl, err = decodeState(state, config.SecretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// load user info
	client := github.NewTokenClient(context.Background(), token.AccessToken)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create installation token
	installationToken, installationID, err := getInstallationToken(context.Background(), user.GetLogin())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, middleware.JWTClaims{
			OAuthToken:        token.AccessToken,
			InstallationToken: installationToken,
			InstallationID:    installationID,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   user.GetLogin(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		},
	)
	session, err := jwtToken.SignedString(config.SecretKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session", session, 3600, "/", "", true, true)
	c.Redirect(http.StatusFound, returnUrl)
}

func getInstallationToken(ctx context.Context, login string) (string, int64, error) {
	// 获取 installationID
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, config.AppID, config.AppPrivateKey)
	if err != nil {
		return "", 0, err
	}
	client := github.NewClient(&http.Client{Transport: atr})
	installation, _, err := client.Apps.FindUserInstallation(ctx, login)
	if err != nil {
		return "", 0, err
	}

	// 生成 installation token
	tr := ghinstallation.NewFromAppsTransport(atr, installation.GetID())
	token, err := tr.Token(ctx)
	if err != nil {
		return "", 0, err
	}
	return token, installation.GetID(), nil
}

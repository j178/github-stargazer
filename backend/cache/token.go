package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v53/github"
	"github.com/redis/rueidis"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/backend/config"
)

type TokenType string

const (
	OAuthTokenType        TokenType = "oauth"
	InstallationTokenType TokenType = "installation"
)

func GetOAuthToken(ctx context.Context, login string) (string, error) {
	var token oauth2.Token
	err := Get(ctx, string(OAuthTokenType), login, &token)
	// 不存在，则无法凭空创建。已存在，则可以根据 refresh_token 刷新
	if err != nil {
		return "", err
	}

	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     oauthGitHub.Endpoint,
	}
	tokenSource := cfg.TokenSource(ctx, &token)
	// reuse or refresh
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", err
	}
	return newToken.AccessToken, nil
}

func SaveOAuthToken(ctx context.Context, login string, token *oauth2.Token) error {
	return Set(ctx, string(OAuthTokenType), login, token, FOREVER)
}

type InstallationToken struct {
	Token          string `json:"token"`
	ExpiresAt      int64  `json:"expires_at"`
	InstallationID int64  `json:"installation_id"`
}

func GetInstallationToken(ctx context.Context, login string) (string, error) {
	var token InstallationToken
	err := Get(ctx, string(InstallationTokenType), login, &token)
	valid := true
	// 不存在，也可以凭空创建
	if rueidis.IsRedisNil(err) {
		valid = false
	} else if err != nil {
		return "", err
	}

	if valid && token.ExpiresAt < time.Now().Unix() {
		valid = false
	}

	if valid {
		return token.Token, nil
	}

	s, installationID, err := createInstallationToken(ctx, login)
	if err != nil {
		return "", err
	}
	token.Token = s
	token.InstallationID = installationID
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app
	// The installation access token will expire after 1 hour.
	token.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()

	err = Set(ctx, string(InstallationTokenType), login, token, FOREVER)
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

func createInstallationToken(ctx context.Context, login string) (string, int64, error) {
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

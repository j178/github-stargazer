package cache

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
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
	// reuse
	if token.Valid() {
		return token.AccessToken, nil
	}

	// refresh and update
	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     oauthGitHub.Endpoint,
	}
	tokenSource := cfg.TokenSource(ctx, &token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", err
	}
	err = Set(ctx, string(OAuthTokenType), login, newToken, FOREVER)
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

func GetInstallationToken(ctx context.Context, installationID int64) (string, error) {
	var token InstallationToken
	installationIDStr := strconv.FormatInt(installationID, 10)
	err := Get(ctx, string(InstallationTokenType), installationIDStr, &token)
	valid := true
	// 不存在，也可以凭空创建
	if err == ErrCacheMiss {
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

	newToken, err := createInstallationToken(ctx, installationID)
	if err != nil {
		return "", err
	}
	token.Token = newToken
	token.InstallationID = installationID
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app
	// The installation access token will expire after 1 hour.
	token.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()

	err = Set(ctx, string(InstallationTokenType), installationIDStr, token, FOREVER)
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

func createInstallationToken(ctx context.Context, installationID int64) (string, error) {
	tr, err := ghinstallation.New(http.DefaultTransport, config.AppID, installationID, config.AppPrivateKey)
	token, err := tr.Token(ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}

package api

import (
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/config"
)

// 提供一个配置的入口

func Authorize(w http.ResponseWriter, r *http.Request) {
	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     github.Endpoint,
	}
	url := cfg.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusFound)
}

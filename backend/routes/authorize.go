package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/backend/config"
	"github.com/j178/github_stargazer/backend/utils"
)

func Authorize(c *gin.Context) {
	returnUrl := c.Query("return_url")
	if returnUrl == "" {
		returnUrl = "/"
	}
	// encrypt return url as state
	state := EncodeState(returnUrl, config.SecretKey)
	origin := fmt.Sprintf("%s://%s", utils.RequestScheme(c), c.Request.Host)
	redirectUrl := fmt.Sprintf("%s/api/authorized", origin)

	cfg := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectUrl,
	}
	url := cfg.AuthCodeURL(state)
	c.Redirect(http.StatusFound, url)
}

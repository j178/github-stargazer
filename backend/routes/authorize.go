package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	state := encodeState(returnUrl, config.SecretKey)
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

func encodeState(state string, secretKey []byte) string {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"state": state,
		},
	)
	token, _ := jwtToken.SignedString(secretKey)
	return token
}

func decodeState(state string, secretKey []byte) (string, error) {
	token, err := jwt.Parse(
		state, func(token *jwt.Token) (any, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["state"].(string), nil
	}
	return "", fmt.Errorf("invalid state")
}

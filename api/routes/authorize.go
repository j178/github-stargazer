package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/j178/github_stargazer/config"
)

func Authorize(c *gin.Context) {
	returnUrl := c.Query("return_url")
	if returnUrl == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "return_url is empty")
		return
	}
	// encrypt return url as state
	state := encodeState(returnUrl, config.SecretKey)
	origin := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.URL.Host)
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

func encodeState(state, secretKey string) string {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"state": state,
		},
	)
	token, _ := jwtToken.SignedString([]byte(secretKey))
	return token
}

func decodeState(state, secretKey string) (string, error) {
	token, err := jwt.Parse(
		state, func(token *jwt.Token) (any, error) {
			return []byte(secretKey), nil
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

package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	OAuthToken        string `json:"access_token"`
	InstallationToken string `json:"installation_token"`
	InstallationID    int64  `json:"installation_id"`
	jwt.RegisteredClaims
}

func (c *JWTClaims) Validate() error {
	if c.OAuthToken == "" || c.InstallationToken == "" || c.InstallationID == 0 {
		return errors.New("invalid claims")
	}

	return nil
}

func CheckJWT(secret []byte) gin.HandlerFunc {
	keyFunc := func(token *jwt.Token) (any, error) {
		return secret, nil
	}

	return func(c *gin.Context) {
		session, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		token, err := jwt.ParseWithClaims(
			session,
			&JWTClaims{},
			keyFunc,
			jwt.WithLeeway(5*time.Second),
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		claims := token.Claims.(*JWTClaims)

		c.Set("jwt", claims)
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CheckJWT(secret []byte) gin.HandlerFunc {
	keyFunc := func(token *jwt.Token) (any, error) {
		return secret, nil
	}

	return func(c *gin.Context) {
		session, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}

		token, err := jwt.Parse(session, keyFunc, jwt.WithLeeway(5*time.Second))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		if claims["login"] == nil || claims["access_token"] == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid session")
			return
		}

		c.Set("login", claims["login"])
		c.Set("access_token", claims["access_token"])
		c.Next()
	}
}

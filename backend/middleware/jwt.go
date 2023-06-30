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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		token, err := jwt.ParseWithClaims(
			session,
			&jwt.RegisteredClaims{},
			keyFunc,
			jwt.WithLeeway(5*time.Second),
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil {
			c.SetCookie("session", "", -1, "/", "", true, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		claims := token.Claims.(*jwt.RegisteredClaims)
		c.Set("login", claims.Subject)
		c.Next()
	}
}

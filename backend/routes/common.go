package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func EncodeState(state string, secretKey []byte) string {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"state": state,
		},
	)
	token, _ := jwtToken.SignedString(secretKey)
	return token
}

func DecodeState(state string, secretKey []byte) (string, error) {
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

func Abort(c *gin.Context, code int, err error, msg string) {
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if err == nil {
		err = errors.New(msg)
	} else {
		err = errors.WithMessage(err, msg)
	}
	c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
}

package middleware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/aki-0517/go-user-management/util"
)

type MiddleWare struct {
	jwtkey []byte
}

func NewMiddleware(jwtkey []byte) *MiddleWare {
	return &MiddleWare{jwtkey}
}

func (m *MiddleWare) AuthenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHader := c.Request.Header.Get("Authorization")
		if authHader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		tokenString, err := util.ExtractBearerToken(authHader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtkey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error parsing token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

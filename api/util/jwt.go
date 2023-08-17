package util

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func GenerateToken(jwtkey []byte, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(jwtKey []byte, tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token")
	}
	return claims, nil
}

func RefreshJWTToken(jwtKey []byte, oldTokenString string) (string, error) {
	oldClaims, err := ParseToken(jwtKey, oldTokenString)
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	newClaims := &jwt.StandardClaims{
		Subject:   oldClaims.Subject,
		ExpiresAt: expirationTime.Unix(),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return newTokenString, nil
}

func GetSubjectFromJWT(c *gin.Context, jwtkey []byte) (string, error) {
	authHader := c.Request.Header.Get("Authorization")
	if authHader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return "", errors.New("Authorization header required")
	}
	tokenString, err := ExtractBearerToken(authHader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return "", err
	}

	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return jwtkey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", errors.New("Invalid token")
}

func AddTokenToBlacklist(token string, rdb *redis.Client, expiration time.Duration) error {
	err := rdb.Set(ctx, token, true, expiration).Err()
	return err
}

func IsTokenBlocklisted(token string, rdb *redis.Client) (bool, error) {
	val, err := rdb.Get(ctx, token).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "true", nil

}

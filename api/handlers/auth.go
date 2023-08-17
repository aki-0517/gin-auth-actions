package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/aki-0517/go-user-management/models"
	"github.com/aki-0517/go-user-management/util"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	JWTKey []byte
	rdb    *redis.Client
}

func AuthHandlerInit(db *gorm.DB, jwtkey []byte, rdb *redis.Client) AuthHandler {
	return AuthHandler{
		db:     db,
		JWTKey: jwtkey,
		rdb:    rdb,
	}
}

func (h *AuthHandler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		foundUser, err := models.GetUserByEmail(h.db, user.Email)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
			return
		} else if foundUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// ユーザーが入力したパスワードと、データベースに保存されているハッシュ化されたパスワードを比較
		if !util.CheckPasswordHash(user.Password, foundUser.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		tokenString, err := util.GenerateToken(h.JWTKey, foundUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

func (h *AuthHandler) ChangePasswordHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.processAndBlacklistToken(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var changePasswordRequest struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := c.ShouldBindJSON(&changePasswordRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		email, err := util.GetSubjectFromJWT(c, h.JWTKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		user, err := models.GetUserByEmail(h.db, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No user found with this email"})
			return
		}

		if !util.CheckPasswordHash(changePasswordRequest.OldPassword, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		hashedPassword, err := util.HashPassword(changePasswordRequest.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = hashedPassword

		if err := h.db.Save(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
	}
}

func (h *AuthHandler) RefreshTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.processAndBlacklistToken(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		oldToken, err := util.ExtractBearerToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		newToken, err := util.RefreshJWTToken(h.JWTKey, oldToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": newToken})
	}
}

func (h *AuthHandler) LogOutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.processAndBlacklistToken(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

func (h *AuthHandler) processAndBlacklistToken(c *gin.Context) error {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return errors.New("Authorization header required")
	}
	tokenString, err := util.ExtractBearerToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return err
	}

	isBlacklisted, err := util.IsTokenBlocklisted(tokenString, h.rdb)
	if err != nil {
		return err
	}

	if isBlacklisted {
		return errors.New("Token is blacklisted")
	}

	claims, err := util.ParseToken(h.JWTKey, tokenString)
	if err != nil {
		return err
	}

	expireTime := time.Unix(claims.ExpiresAt, 0)
	expiration := expireTime.Sub(time.Now())

	return util.AddTokenToBlacklist(tokenString, h.rdb, expiration)
}

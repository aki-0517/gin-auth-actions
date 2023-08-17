package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/aki-0517/go-user-management/models"
	"github.com/aki-0517/go-user-management/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	JWTKey []byte
}

func UserHandler(db *gorm.DB, jwtKey []byte) *Handler {
	return &Handler{
		db:     db,
		JWTKey: jwtKey,
	}
}

func getUUIDFromRequest(c *gin.Context) uuid.UUID {
	id := c.Param("id")
	uid, err := uuid.Parse(id)

	if err != nil {
		panic(err)
	}
	return uid
}

func (h *Handler) ListUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := models.GetAllUsers(h.db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func (h *Handler) GetUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := getUUIDFromRequest(c)
		user, err := models.GetUserById(h.db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func (h *Handler) CreateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User

		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user.Name == "" || user.Email == "" || user.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		newUser, err := models.CreateUser(h.db, user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "user created" + newUser.Name})
	}
}

func (h *Handler) UpdateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := getUUIDFromRequest(c)

		currentUser, err := models.GetUserById(h.db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var updatedInfo models.User
		if err := c.ShouldBind(&updatedInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if updatedInfo.Name != "" {
			currentUser.Name = updatedInfo.Name
		}
		if updatedInfo.Email != "" {
			currentUser.Email = updatedInfo.Email
		}

		updatedUser, err := models.UpdateUser(h.db, *currentUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tokenString, err := util.GenerateToken(h.JWTKey, updatedUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token":   tokenString,
			"user":    updatedUser,
			"message": "User updated successfully",
		})
	}
}

func (h *Handler) DeleteUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := getUUIDFromRequest(c)

		user, err := models.GetUserById(h.db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		isDeleted, err := models.DeleteUser(h.db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isDeleted {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
	}
}

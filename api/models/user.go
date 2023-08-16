package models

import (
	"errors"

	"github.com/google/uuid"
	"github.com/aki-0517/go-user-management/util"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

type DBConfig struct {
	HOST     string
	PORT     string
	USER     string
	DBNAME   string
	PASSWORD string
}

func (u *User) IsEqual(user *User) bool {
	if u.ID != user.ID {
		return false
	}
	if u.Name != user.Name {
		return false
	}
	if u.Email != user.Email {
		return false
	}
	if u.Password != user.Password {
		return false
	}
	return true
}

func isEmailUnique(db *gorm.DB, email string) bool {
	isUserUnique, err := GetUserByEmail(db, email)
	if err != nil || isUserUnique == nil {
		return false
	}
	return true
}

func CreateUser(db *gorm.DB, user User) (*User, error) {

	if user.Name == "" || user.Email == "" || user.Password == "" {
		return nil, errors.New("name, email and password are required")
	}

	if isEmailUnique(db, user.Email) {
		return nil, errors.New("email is already used")
	}

	hashedPassword, err := util.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword

	result := db.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func GetUserById(db *gorm.DB, id uuid.UUID) (*User, error) {
	var user User
	result := db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, user User) (*User, error) {
	result := db.Save(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func DeleteUser(db *gorm.DB, user *User) (bool, error) {
	result := db.Delete(&user)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

package models

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "password"
	dbname := "postgres"
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error
	if err != nil {
		panic("failed to create pgcrypto extension: " + err.Error())
	}

	db.AutoMigrate(&User{})

	return db
}

func teardownTestDB(db *gorm.DB) {
	db.Migrator().DropTable(&User{})

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get sqlDB")
	}
	sqlDB.Close()
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	user := User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "test",
	}

	createdUser, err := CreateUser(db, user)
	assert.Nil(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, user.Name, createdUser.Name)

	emptyEmailUser := User{
		Name:     "test",
		Email:    "",
		Password: "test",
	}
	createdEmptyEmailUser, err := CreateUser(db, emptyEmailUser)
	assert.NotNil(t, err)
	assert.Nil(t, createdEmptyEmailUser)

	emptyPasswordUser := User{
		Name:     "test",
		Email:    "test1@test.com",
		Password: "",
	}
	createdEmptyPasswordUser, err := CreateUser(db, emptyPasswordUser)
	assert.NotNil(t, err)
	assert.Nil(t, createdEmptyPasswordUser)

	emptyNameUser := User{
		Name:     "",
		Email:    "test2@test.com",
		Password: "test",
	}
	createdEmptyNameUser, err := CreateUser(db, emptyNameUser)
	assert.NotNil(t, err)
	assert.Nil(t, createdEmptyNameUser)

	duplicatedEmailUser := User{
		Name:     "test1",
		Email:    "test@test.com",
		Password: "test1",
	}
	createdDuplicatedEmailUser, err := CreateUser(db, duplicatedEmailUser)
	assert.NotNil(t, err)
	assert.Nil(t, createdDuplicatedEmailUser)
}

func TestGetUserById(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	user := User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "test",
	}

	createdUser, err := CreateUser(db, user)

	gotUser, err := GetUserById(db, createdUser.ID)
	assert.Nil(t, err)
	assert.NotNil(t, gotUser)
	assert.True(t, createdUser.IsEqual(gotUser))

	gotUserWithWrongId, err := GetUserById(db, uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, gotUserWithWrongId)
}

func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	user := User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "test",
	}

	createdUser, err := CreateUser(db, user)

	gotUser, err := GetUserByEmail(db, createdUser.Email)
	assert.Nil(t, err)
	assert.NotNil(t, gotUser)
	assert.True(t, createdUser.IsEqual(gotUser))

	gotUserWithWrongEmail, err := GetUserByEmail(db, "")
	assert.Nil(t, err)
	assert.Nil(t, gotUserWithWrongEmail)
}

func TestGetAllUsers(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	emptyUsers, err := GetAllUsers(db)
	assert.Nil(t, err)
	assert.NotNil(t, emptyUsers)
	assert.Equal(t, 0, len(emptyUsers))

	user1 := User{
		Name:     "test1",
		Email:    "test1@test.com",
		Password: "test1",
	}
	user2 := User{
		Name:     "test2",
		Email:    "test2@test.com",
		Password: "test2",
	}

	createdUser1, err := CreateUser(db, user1)

	createdUser2, err := CreateUser(db, user2)

	gotUsers, err := GetAllUsers(db)
	assert.Nil(t, err)
	assert.NotNil(t, gotUsers)
	assert.Equal(t, 2, len(gotUsers))
	assert.True(t, createdUser1.IsEqual(&gotUsers[0]))
	assert.True(t, createdUser2.IsEqual(&gotUsers[1]))
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	user := User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "test",
	}

	createdUser, err := CreateUser(db, user)

	createdUser.Name = "test2"
	updatedUser, err := UpdateUser(db, *createdUser)
	assert.Nil(t, err)
	assert.NotNil(t, updatedUser)
	assert.True(t, createdUser.IsEqual(updatedUser))
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB()
	defer teardownTestDB(db)

	user := User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "test",
	}

	createdUser, err := CreateUser(db, user)

	isDeleted, err := DeleteUser(db, createdUser)
	assert.Nil(t, err)
	assert.True(t, isDeleted)

	gotUser, err := GetUserById(db, createdUser.ID)
	assert.Nil(t, gotUser)
	assert.Nil(t, err)

	deleted, err := DeleteUser(db, createdUser)
	assert.Nil(t, err)
	assert.False(t, deleted)
}

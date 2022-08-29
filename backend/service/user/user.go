package user

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

type User entity.User

func GetByEmail(email string) (User, error) {
	db := db.GetDB()
	var u User
	if err := db.Where("email = ?", email).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

func GetByUsername(username string) (User, error) {
	db := db.GetDB()
	var u User
	if err := db.Where("name = ?", username).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

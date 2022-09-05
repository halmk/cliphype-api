package user

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

func Create(username, email string) (entity.User, error) {
	db := db.GetDB()
	var u entity.User
	{
		u.Name = username
		u.Email = email
		u.IsActive = false
		u.IsStaff = false
		u.IsSuperuser = false
	}
	if err := db.Create(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

func GetByID(id int) (entity.User, error) {
	db := db.GetDB()
	var u entity.User
	if err := db.Where("id = ?", id).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

func GetByEmail(email string) (entity.User, error) {
	db := db.GetDB()
	var u entity.User
	if err := db.Where("email = ?", email).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

func GetByUsername(username string) (entity.User, error) {
	db := db.GetDB()
	var u entity.User
	if err := db.Where("name = ?", username).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

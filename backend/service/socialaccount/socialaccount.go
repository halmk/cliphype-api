package socialaccount

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

type Socialaccount entity.Socialaccount

func GetByUserId(user_id uint) (Socialaccount, error) {
	db := db.GetDB()
	var sa Socialaccount
	if err := db.Where("user_id = ?", user_id).First(&sa).Error; err != nil {
		return sa, err
	}
	return sa, nil
}

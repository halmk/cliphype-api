package socialaccount

import (
	"encoding/json"

	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func Create(user entity.User, provider entity.Provider, info map[string]interface{}) (entity.Socialaccount, error) {
	db := db.GetDB()
	var sa entity.Socialaccount
	{
		sa.User = user
		sa.Provider = provider
		info_bytes, err := json.Marshal(info)
		if err != nil {
			return sa, err
		}
		sa.ExtraData = postgres.Jsonb{RawMessage: info_bytes}
	}
	if err := db.Create(&sa).Error; err != nil {
		return sa, err
	}

	return sa, nil
}
func GetByUserId(user_id uint) (entity.Socialaccount, error) {
	db := db.GetDB()
	var sa entity.Socialaccount
	if err := db.Where("user_id = ?", user_id).First(&sa).Error; err != nil {
		return sa, err
	}
	return sa, nil
}

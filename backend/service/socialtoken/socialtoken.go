package socialtoken

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

type Socialtoken entity.Socialtoken

func GetBySocialaccountId(socialaccount_id uint) (Socialtoken, error) {
	db := db.GetDB()
	var st Socialtoken
	if err := db.Where("socialaccount_id = ?", socialaccount_id).First(&st).Error; err != nil {
		return st, err
	}
	return st, nil
}

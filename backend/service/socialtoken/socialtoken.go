package socialtoken

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
	"golang.org/x/oauth2"
)

func Create(sa entity.Socialaccount, provider entity.Provider, token *oauth2.Token) (entity.Socialtoken, error) {
	db := db.GetDB()
	var st entity.Socialtoken
	{
		st.Provider = provider
		st.Socialaccount = sa
		st.AccessToken = token.AccessToken
		st.RefreshToken = token.RefreshToken
		st.Expiry = token.Expiry
	}
	if err := db.Create(&st).Error; err != nil {
		return st, err
	}

	return st, nil
}
func GetBySocialaccountId(socialaccount_id uint) (entity.Socialtoken, error) {
	db := db.GetDB()
	var st entity.Socialtoken
	if err := db.Where("socialaccount_id = ?", socialaccount_id).First(&st).Error; err != nil {
		return st, err
	}
	return st, nil
}

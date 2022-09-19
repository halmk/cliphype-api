package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Socialtoken struct {
	gorm.Model
	ProviderID      int
	Provider        Provider
	SocialaccountID int
	Socialaccount   Socialaccount
	AccessToken     string
	RefreshToken    string
	Expiry          time.Time
}

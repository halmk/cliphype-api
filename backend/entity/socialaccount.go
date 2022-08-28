package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Socialaccount struct {
	gorm.Model
	UserID     int
	User       User
	ProviderID int
	Provider   Provider
	LastLogin  time.Time
	ExtraData  postgres.Jsonb
}

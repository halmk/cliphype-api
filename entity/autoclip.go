package entity

import (
	"github.com/jinzhu/gorm"
)

type AutoClip struct {
	gorm.Model
	UserID   int
	User     User
	Streamer string
	ClipID   string
	EditURL  string
	Hype     float64
}

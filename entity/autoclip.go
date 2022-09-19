package entity

import (
	"github.com/jinzhu/gorm"
)

type AutoClip struct {
	gorm.Model
	UserID  int
	User    User
	ClipID  string
	EditURL string
}

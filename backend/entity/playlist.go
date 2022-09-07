package entity

import (
	"github.com/jinzhu/gorm"
)

type Playlist struct {
	gorm.Model
	Title      string
	StreamerID int
	Streamer   Streamer
	CreatorID  int
	Creator    User
	VideoID    string
}

package entity

import (
	"github.com/jinzhu/gorm"
)

type PlaylistClip struct {
	gorm.Model
	PlaylistID int
	Playlist   Playlist
	ClipID     string
}

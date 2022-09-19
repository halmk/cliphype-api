package entity

import (
	"github.com/jinzhu/gorm"
)

type PlaylistClip struct {
	gorm.Model
	PlaylistID   int
	Playlist     Playlist
	ClipID       string
	Duration     float64
	EmbedURL     string
	ThumbnailURL string
	Title        string
	URL          string
	VideoID      string
	VodOffset    int
}

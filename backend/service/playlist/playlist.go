package playlist

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

func Create(title string, streamer entity.Streamer, creator entity.User) (entity.Playlist, error) {
	db := db.GetDB()
	var playlist entity.Playlist
	{
		playlist.Title = title
		playlist.Streamer = streamer
		playlist.Creator = creator
	}
	if err := db.Create(&playlist).Error; err != nil {
		return playlist, err
	}

	return playlist, nil
}

func GetAll() ([]entity.Playlist, error) {
	db := db.GetDB()
	var playlists []entity.Playlist
	if err := db.Find(&playlists).Error; err != nil {
		return playlists, err
	}
	return playlists, nil
}

func GetByStreamerID(streamer_id uint) ([]entity.Playlist, error) {
	db := db.GetDB()
	var playlists []entity.Playlist
	if err := db.Where("streamer_id = ?", streamer_id).Find(&playlists).Error; err != nil {
		return playlists, err
	}
	return playlists, nil
}

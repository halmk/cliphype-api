package playlist

import (
	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/entity"
)

func Create(title string, streamer entity.Streamer, creator *entity.User, video_id *string) (entity.Playlist, error) {
	db := db.GetDB()
	var playlist entity.Playlist
	{
		playlist.Title = title
		playlist.Streamer = streamer
		if creator != nil {
			playlist.Creator = *creator
		}
		if video_id != nil {
			playlist.VideoID = *video_id
		}
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

func GetByVideoID(video_id string) ([]entity.Playlist, error) {
	db := db.GetDB()
	var playlists []entity.Playlist
	if err := db.Where("video_id = ?", video_id).Find(&playlists).Error; err != nil {
		return playlists, err
	}
	return playlists, nil
}

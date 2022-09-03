package playlistclip

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

func Create(clip_id string, playlist entity.Playlist) (entity.PlaylistClip, error) {
	db := db.GetDB()
	var playlist_clip entity.PlaylistClip
	{
		playlist_clip.ClipID = clip_id
		playlist_clip.Playlist = playlist
	}
	if err := db.Create(&playlist_clip).Error; err != nil {
		return playlist_clip, err
	}

	return playlist_clip, nil
}

func GetAll() ([]entity.PlaylistClip, error) {
	db := db.GetDB()
	var playlist_clips []entity.PlaylistClip
	if err := db.Find(&playlist_clips).Error; err != nil {
		return playlist_clips, err
	}
	return playlist_clips, nil
}

func GetArrayByPlaylist(playlist_id uint) ([]entity.PlaylistClip, error) {
	db := db.GetDB()
	var playlist_clips []entity.PlaylistClip
	if err := db.Where("playlist_id = ?", playlist_id).Find(&playlist_clips).Error; err != nil {
		return playlist_clips, err
	}
	return playlist_clips, nil
}

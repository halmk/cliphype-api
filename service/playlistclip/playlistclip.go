package playlistclip

import (
	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/entity"
)

func Create(clip_id string, duration float64, embed_url string, thumbnail_url string, title string, url string, video_id string, vod_offset int, playlist entity.Playlist) (entity.PlaylistClip, error) {
	db := db.GetDB()
	var playlist_clip entity.PlaylistClip
	{
		playlist_clip.ClipID = clip_id
		playlist_clip.Duration = duration
		playlist_clip.EmbedURL = embed_url
		playlist_clip.ThumbnailURL = thumbnail_url
		playlist_clip.Title = title
		playlist_clip.URL = url
		playlist_clip.VideoID = video_id
		playlist_clip.VodOffset = vod_offset
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

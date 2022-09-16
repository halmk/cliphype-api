package autoclip

import (
	"github.com/halmk/cliplist-ttv/backend/db"
	"github.com/halmk/cliplist-ttv/backend/entity"
)

func Create(clip_id string, edit_url string, user entity.User) (entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clip entity.AutoClip
	{
		auto_clip.ClipID = clip_id
		auto_clip.EditURL = edit_url
		auto_clip.User = user
	}
	if err := db.Create(&auto_clip).Error; err != nil {
		return auto_clip, err
	}

	return auto_clip, nil
}

func GetAll() ([]entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clips []entity.AutoClip
	if err := db.Find(&auto_clips).Error; err != nil {
		return auto_clips, err
	}
	return auto_clips, nil
}

func GetArrayByUser(user_id uint) ([]entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clips []entity.AutoClip
	if err := db.Where("user_id = ?", user_id).Find(&auto_clips).Error; err != nil {
		return auto_clips, err
	}
	return auto_clips, nil
}

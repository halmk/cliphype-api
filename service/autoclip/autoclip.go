package autoclip

import (
	"log"

	"github.com/halmk/cliphype-api/db"
	"github.com/halmk/cliphype-api/entity"
)

func Create(clip_id string, edit_url string, user entity.User, streamer string, hype float64) (entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clip entity.AutoClip
	{
		auto_clip.ClipID = clip_id
		auto_clip.EditURL = edit_url
		auto_clip.User = user
		auto_clip.Streamer = streamer
		auto_clip.Hype = hype
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

func GetArrayBy(user_id *uint, streamer *string) ([]entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clips []entity.AutoClip
	cond := make(map[string]interface{})
	if user_id != nil {
		cond["user_id"] = *user_id
	}
	if streamer != nil {
		cond["streamer"] = *streamer
	}
	if err := db.Where(cond).Find(&auto_clips).Error; err != nil {
		return auto_clips, err
	}
	return auto_clips, nil
}

func GetArrayWhere(query interface{}, args ...interface{}) ([]entity.AutoClip, error) {
	db := db.GetDB()
	var auto_clips []entity.AutoClip
	if err := db.Where(query, args...).Find(&auto_clips).Error; err != nil {
		log.Println(err)
		return auto_clips, err
	}
	return auto_clips, nil
}
